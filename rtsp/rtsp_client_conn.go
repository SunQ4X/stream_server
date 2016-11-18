package rtsp

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/stream_server/media"
)

type RtspClientConnection struct {
	conn net.Conn
}

func NewRtspClientConnection(conn net.Conn) *RtspClientConnection {
	return &RtspClientConnection{conn: conn}
}

func (conn *RtspClientConnection) Handle() {
	fmt.Printf("------ rtsp client connection[%s] : handling ------\n", conn.conn.RemoteAddr())

	for {
		req, err := ReadRequest(conn.conn)
		if err != nil {
			break
		}

		fmt.Printf("------ rtsp client connection[%s] : get request ------ \n%s\n", conn.conn.RemoteAddr(), req)

		//处理RTSP请求
		resp := conn.handleRequestAndReturnResponse(req)

		conn.conn.Write([]byte(resp.String()))
		fmt.Printf("------ Session[%s] : set response ------ \n%s\n", conn.conn.RemoteAddr(), resp)
	}

	conn.conn.Close()
	fmt.Printf("------ Session[%s] : closed ------\n", conn.conn.RemoteAddr())
}

func (conn *RtspClientConnection) handleRequestAndReturnResponse(req *Request) *Response {
	cSeq := req.Header.Get("CSeq")

	reqUrl, err := url.ParseRequestURI(req.URL)
	if err != nil {
		return NewResponse(BadGateway, "Url parse error", cSeq, "")
	}

	switch req.Method {
	case OPTIONS:
		resp := NewResponse(OK, "OK", cSeq, "")
		options := strings.Join([]string{OPTIONS, DESCRIBE, SETUP, PLAY, TEARDOWN}, ", ")
		resp.Header["Public"] = []string{options}
		return resp

	case DESCRIBE:
		query := reqUrl.Query()
		serialNum := query.Get("sn")
		chNo := query.Get("ch")
		streamType := query.Get("stream_type")
		if "" == serialNum || "" == chNo || "" == streamType {
			return NewResponse(BadRequest, "Stream parameter error", cSeq, "")
		}

		streamName := fmt.Sprintf("%s&%s&%s", serialNum, chNo, streamType)
		mediaSess, exits := media.LookupMediaSession(streamName)
		if !exits {
			return NewResponse(SessionNotFound, "Session not found", cSeq, "")
		}

		sdp := mediaSess.GenerateSDPDescription()

		resp := NewResponse(OK, "OK", cSeq, "")
		resp.Header.Add("Content-Base", req.URL)
		resp.Header.Add("Content-Type", "application/sdp")
		resp.Header.Add("Content-Length", strconv.Itoa(len(sdp)))
		resp.Body = sdp

		return resp

	case SETUP:

	case PLAY:

	case TEARDOWN:

	default:
		return NewResponse(MethodNotAllowed, "Option Not Support", cSeq, "")
	}

	return NewResponse(OK, "OK", cSeq, "")
}
