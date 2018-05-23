package rtsp_server

import (
	"crypto/rand"
	"encoding/base64"
	//	"encoding/json"
	"fmt"
	"io"
	"logger"
	"logic_proc"
	"media_session"
	"net"
	"net/url"
	//	"protocol"
	"rtsp"
	"strconv"
	"stream"
	"strings"
	//	"utility"

	//"github.com/garyburd/redigo/redis"
)

var (
	RTP_TRANSPORT_MODE   = [2]string{"RTP/AVP/TCP", "RTP/AVP"}
	RTP_CAST_MODE        = [2]string{"unicast", "multicast"}
	STREAM_HEADER_LENGTH = 4
)

type RtspClientConnection struct {
	conn         net.Conn
	sessionID    string
	mediaSession *media_session.MediaSession
	processor    *logic_proc.Processor
}

func NewRtspClientConnection(conn net.Conn) *RtspClientConnection {
	b := make([]byte, 32)
	io.ReadFull(rand.Reader, b)
	sessionID := base64.URLEncoding.EncodeToString(b)
	processor := logic_proc.GetProcessor()
	return &RtspClientConnection{
		conn:      conn,
		sessionID: sessionID,
		processor: processor,
	}
}

func (conn *RtspClientConnection) Handle() {
	logger.Debug("------ rtsp client connection[", conn.conn.RemoteAddr(), "] : handling ------")

	defer func() {
		if re := recover(); re != nil {
			logger.Error("RtspClientConnection Handle panic:", re)
		}

		conn.processor.RemoveClient(conn.conn.RemoteAddr().String())

		conn.conn.Close()

		if conn.mediaSession != nil {
			conn.mediaSession.Close()
			conn.mediaSession = nil
		}

		logger.Debug("------ connection[", conn.conn.RemoteAddr(), "] : closed ------")
	}()

	buffer := make([]byte, 2048)
	length := 0

	for {
		recvLen, err := conn.conn.Read(buffer[length:])
		if err != nil {
			//logger.Error("conn read data error:", err)
			return
		}

		length += recvLen

		if buffer[0] == '$' {
			//			logger.Debug("stream data!")

			for length < STREAM_HEADER_LENGTH {
				recvLen, err := conn.conn.Read(buffer[length:])
				if err != nil {
					logger.Error("conn read data error:", err)
					return
				}

				length += recvLen
			}

			tcpChannel := int(buffer[1])
			streamDataLength := ((int(buffer[2]) << 8) | int(buffer[3]))

			streamDataRecvLength := length - STREAM_HEADER_LENGTH

			for streamDataRecvLength < streamDataLength {
				recvLen, err := conn.conn.Read(buffer[length:])
				if err != nil {
					logger.Error("conn read data error:", err)
					return
				}

				length += recvLen
				streamDataRecvLength = length - STREAM_HEADER_LENGTH
			}

			dataBuffer := make([]byte, streamDataLength)
			copy(dataBuffer, buffer[STREAM_HEADER_LENGTH:STREAM_HEADER_LENGTH+streamDataLength])
			length = copy(buffer, buffer[STREAM_HEADER_LENGTH+streamDataLength:length])

			if conn.mediaSession != nil {
				conn.mediaSession.HandleStreamData(tcpChannel, dataBuffer)
			}
		} else {
			recv := string(buffer[:length])

			req, err := rtsp.ReadRequest(recv)
			if err != nil {
				logger.Error("rtsp read request error:", err)
				return
			}

			logger.Debug("------ rtsp client connection[", conn.conn.RemoteAddr(), "] : get request ------ \r\n", req)

			//处理RTSP请求
			var resp *rtsp.Response
			switch req.Method {
			case rtsp.OPTIONS:
				resp = conn.handleOptions(req)
			case rtsp.DESCRIBE:
				resp = conn.handleDescribe(req)
			case rtsp.SETUP:
				resp = conn.handleSetup(req)
			case rtsp.PLAY:
				resp = conn.handlePlay(req)
			case rtsp.TEARDOWN:
				resp = conn.handleTeardown(req)
			case rtsp.GET_PARAMETER:
				resp = conn.handleGetParameter(req)
			case rtsp.SET_PARAMETER:
				resp = conn.handleSetParameter(req)
			default:
				resp = rtsp.NewResponse(rtsp.MethodNotAllowed, "Method Not Allowed", req.Header["CSeq"], "")
			}

			conn.conn.Write([]byte(resp.String()))

			logger.Debug("------ Session[", conn.conn.RemoteAddr(), "] : set response ------ \r\n", resp)

			length = 0
		}
	}
}

func (conn *RtspClientConnection) handleOptions(req *rtsp.Request) *rtsp.Response {
	resp := rtsp.NewResponse(rtsp.OK, "OK", req.Header["CSeq"], "")
	resp.Header["Public"] = strings.Join([]string{rtsp.OPTIONS, rtsp.DESCRIBE, rtsp.SETUP, rtsp.PLAY, rtsp.PAUSE, rtsp.GET_PARAMETER, rtsp.SET_PARAMETER, rtsp.TEARDOWN}, ", ")
	return resp
}

func (conn *RtspClientConnection) handleDescribe(req *rtsp.Request) *rtsp.Response {
	auth, ok := req.Header["Authorization"]
	if !ok || !conn.authorize(auth) {
		resp := rtsp.NewResponse(rtsp.Unauthorized, "Unauthorized", req.Header["CSeq"], "")
		resp.Header["WWW-Authenticate"] = "Basic realm=\"@\""
		return resp
	}

	reqUrl, err := url.Parse(req.URL)
	if err != nil {
		return rtsp.NewResponse(rtsp.BadGateway, err.Error(), req.Header["CSeq"], "")
	}

	streamName := strings.TrimPrefix(reqUrl.RequestURI(), "/")

	sess := media_session.NewMediaSession(streamName)
	if sess == nil {
		return rtsp.NewResponse(rtsp.SessionNotFound, "Session Not Found", req.Header["CSeq"], "")
	}

	conn.mediaSession = sess

	sdplines, err := conn.mediaSession.SDPLines()
	if err != nil {
		return rtsp.NewResponse(rtsp.RequestTimeout, err.Error(), req.Header["CSeq"], "")
	}

	begin, end := conn.mediaSession.Stream.Period()

	var npt string
	if begin < end {
		npt = fmt.Sprintf("0-%d", end-begin)
	} else {
		npt = "0-"
	}

	sdp := fmt.Sprintf(`v=0
o=- %s %s IN IP4 %s
s=Media by ClearView Stream Server
a=control:*
a=range:npt=%s
t=%d %d
%s`,
		conn.sessionID, conn.sessionID, conn.conn.LocalAddr().String(),
		npt,
		begin, end,
		sdplines)

	resp := rtsp.NewResponse(rtsp.OK, "OK", req.Header["CSeq"], sdp)
	resp.Header["Content-Base"] = req.URL
	resp.Header["Content-Type"] = "application/sdp"
	resp.Header["Content-Length"] = strconv.Itoa(len(sdp))

	conn.processor.AddClient(conn.conn.RemoteAddr().String(), streamName)

	return resp
}

func (conn *RtspClientConnection) handleSetup(req *rtsp.Request) *rtsp.Response {
	var trackId string

	trackPos := strings.LastIndex(req.URL, conn.mediaSession.StreamName+"/")
	if -1 == trackPos {
		trackId = ""
	} else {
		trackId = string([]byte(req.URL)[trackPos+len(conn.mediaSession.StreamName)+1:])
	}

	subSession := conn.mediaSession.LookupSubSession(trackId)
	if subSession == nil {
		return rtsp.NewResponse(rtsp.SessionNotFound, "Sub Session Not Found", req.Header["CSeq"], "")
	}

	transport := req.Header["Transport"]

	transportMode, castMode, rtpPort, rtcpPort, rtpChannelId, rtcpChannelId := parseReqTransport(transport)

	if castMode == media_session.MULTICAST {
		return rtsp.NewResponse(rtsp.UnsupportedTransport, "Unsupported Transport", req.Header["CSeq"], "")
	}

	if transportMode != media_session.RTP_TCP {
		return rtsp.NewResponse(rtsp.UnsupportedTransport, "Unsupported Transport", req.Header["CSeq"], "")
	}

	dest := media_session.NewDestination(transportMode, castMode, rtpPort, rtcpPort, rtpChannelId, rtcpChannelId, conn.conn)
	if dest == nil {
		return rtsp.NewResponse(rtsp.DestinationUnreachable, "Destination Unreachable", req.Header["CSeq"], "")
	}

	subSession.SetDestination(dest)

	var respTransport string
	if transportMode == media_session.RTP_UDP {
		respTransport = fmt.Sprintf("%s;%s;client_port=%d-%d;server_port=%d-%d",
			RTP_TRANSPORT_MODE[transportMode],
			RTP_CAST_MODE[castMode],
			rtpPort,
			rtcpPort,
			dest.LocalRtpPort,
			dest.LocalRtcpPort)
	} else {
		respTransport = fmt.Sprintf("%s;%s;interleaved=%d-%d;ssrc=%d",
			RTP_TRANSPORT_MODE[transportMode],
			RTP_CAST_MODE[castMode],
			rtpChannelId,
			rtcpChannelId,
			subSession.SyncSource())
	}

	resp := rtsp.NewResponse(rtsp.OK, "OK", req.Header["CSeq"], "")
	resp.Header["Session"] = conn.sessionID + "; timeout=60"
	resp.Header["Transport"] = respTransport

	return resp
}

func (conn *RtspClientConnection) handlePlay(req *rtsp.Request) *rtsp.Response {
	stringRange, ok := req.Header["Range"]
	if ok &&
		(conn.mediaSession.Stream.Type() == stream.STREAM_TYPE_DEVICE_REPLAY_I8 ||
			conn.mediaSession.Stream.Type() == stream.STREAM_TYPE_STORAGE_REPLAY_I8 ||
			conn.mediaSession.Stream.Type() == stream.STREAM_TYPE_DEVICE_REPLAY_STANDARD ||
			conn.mediaSession.Stream.Type() == stream.STREAM_TYPE_STORAGE_REPLAY_STANDARD) {
		//重定位时间
		var year, month, day, hour, minute, second, timeZone int

		_, err := fmt.Sscanf(stringRange, "clock=%04d%02d%02dT%02d%02d%02d.%dZ-", &year, &month, &day, &hour, &minute, &second, &timeZone)
		if err != nil {
			logger.Debug("获取定位时间出错:", err, "range:", stringRange)
			return rtsp.NewResponse(rtsp.InvalidRange, "Invalid Range", req.Header["CSeq"], "")
		}

		switch conn.mediaSession.Stream.(type) {
		case *stream.DeviceReplayStream:
			if err := conn.mediaSession.Stream.(*stream.DeviceReplayStream).SeekTime(year, month, day, hour, minute, second, timeZone); err != nil {
				return rtsp.NewResponse(rtsp.ServiceUnavailable, err.Error(), req.Header["CSeq"], "")
			}
		case *stream.StorageReplayStream:
			if err := conn.mediaSession.Stream.(*stream.StorageReplayStream).SeekTime(year, month, day, hour, minute, second, timeZone); err != nil {
				return rtsp.NewResponse(rtsp.ServiceUnavailable, err.Error(), req.Header["CSeq"], "")
			}
		default:
			return rtsp.NewResponse(rtsp.MethodNotValidInThisState, "Method Not Valid In This State", req.Header["CSeq"], "")
		}
	} else {
		//开始播放
		var trackId string

		trackPos := strings.LastIndex(req.URL, conn.mediaSession.StreamName+"/")
		if -1 == trackPos {
			trackId = ""
		} else {
			trackId = string([]byte(req.URL)[trackPos+len(conn.mediaSession.StreamName)+1:])
		}

		_, err := conn.mediaSession.StartPlay(conn.sessionID, trackId)
		if err != nil {
			logger.Debug("start play err:", err)
			return rtsp.NewResponse(rtsp.SessionNotFound, err.Error(), req.Header["CSeq"], "")
		}
	}

	resp := rtsp.NewResponse(rtsp.OK, "OK", req.Header["CSeq"], "")
	resp.Header["Session"] = conn.sessionID

	return resp
}

func (conn *RtspClientConnection) handlePause(req *rtsp.Request) *rtsp.Response {
	if conn.mediaSession.Stream.Type() != stream.STREAM_TYPE_DEVICE_REPLAY_I8 &&
		conn.mediaSession.Stream.Type() != stream.STREAM_TYPE_STORAGE_REPLAY_I8 &&
		conn.mediaSession.Stream.Type() != stream.STREAM_TYPE_DEVICE_REPLAY_STANDARD &&
		conn.mediaSession.Stream.Type() != stream.STREAM_TYPE_STORAGE_REPLAY_STANDARD {
		return rtsp.NewResponse(rtsp.MethodNotAllowed, "Method Not Allowed", req.Header["CSeq"], "")
	}

	return rtsp.NewResponse(rtsp.OK, "OK", req.Header["CSeq"], "")
}

func (conn *RtspClientConnection) handleTeardown(req *rtsp.Request) *rtsp.Response {
	conn.mediaSession.Close()

	conn.mediaSession = nil

	return rtsp.NewResponse(rtsp.OK, "OK", req.Header["CSeq"], "")
}

func (conn *RtspClientConnection) handleGetParameter(req *rtsp.Request) *rtsp.Response {
	resp := rtsp.NewResponse(rtsp.OK, "OK", req.Header["CSeq"], "")
	resp.Header["Session"] = conn.sessionID
	resp.Header["Content-type"] = "text/parameters"
	resp.Header["Content-length"] = "0"
	return resp
}

func (conn *RtspClientConnection) handleSetParameter(req *rtsp.Request) *rtsp.Response {
	resp := rtsp.NewResponse(rtsp.OK, "OK", req.Header["CSeq"], "")
	resp.Header["Session"] = conn.sessionID
	resp.Header["Content-type"] = "text/parameters"
	resp.Header["Content-length"] = "0"
	return resp
}

func parseReqTransport(transport string) (int, int, int, int, int, int) {
	transportMode := media_session.RTP_UDP
	castMode := media_session.UNICAST
	rtpPort := 0
	rtcpPort := 0
	rtpChannelId := 0
	rtcpChannelId := 0

	parts := strings.Split(transport, ";")

	for _, part := range parts {
		if part == "RTP/AVP" {
			transportMode = media_session.RTP_UDP
		} else if part == "RTP/AVP/TCP" {
			transportMode = media_session.RTP_TCP
		} else if part == "unicast" {
			castMode = media_session.UNICAST
		} else if part == "multicast" {
			castMode = media_session.MULTICAST
		} else {
			_, err := fmt.Fscanf(strings.NewReader(part), "client_port=%d-%d", &rtpPort, &rtcpPort)

			if err != nil {
				fmt.Fscanf(strings.NewReader(part), "interleaved=%d-%d", &rtpChannelId, &rtcpChannelId)
			}
		}
	}

	return transportMode, castMode, rtpPort, rtcpPort, rtpChannelId, rtcpChannelId
}

func (conn *RtspClientConnection) authorize(auth string) bool {
	username, password, err := conn.processor.GetRtspAccount()
	if err != nil {
		logger.Error("获取用户认证信息出错", err)

		return false
	}

	auths := strings.SplitN(auth, " ", 2)

	if strings.ToLower(auths[0]) != "basic" {
		return false
	}

	clientAccount, _ := base64.URLEncoding.DecodeString(auths[1])
	clientAccounts := strings.SplitN(string(clientAccount), ":", 2)
	if clientAccounts[0] != username || clientAccounts[1] != password {
		logger.Error("用户认证失败", string(clientAccount))

		return false
	}

	return true
}
