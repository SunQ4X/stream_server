package media_session

import (
	"fmt"
	"io"
	"logger"
	"net"
	"rtcp"
	"rtp"
	"strings"
	"time"
)

const (
	RTP_BEGIN_SOCK_PORT = /*6970*/ 6990
)

const (
	RTP_TCP = 0
	RTP_UDP = 1
)

const (
	UNICAST   = 0
	MULTICAST = 1
)

type Destination struct {
	TransportMode    int
	CastMode         int
	LocalRtpPort     int
	LocalRtcpPort    int
	RtpChannelId     int
	RtcpChannelId    int
	tcpConn          net.Conn
	rtpConn          *net.UDPConn
	rtcpConn         *net.UDPConn
	remoteRtpAddr    *net.UDPAddr
	remoteRtcpAddr   *net.UDPAddr
	isPlaying        bool
	currentSeq       uint
	currentTimestamp uint
	syncSource       uint
	sendPacketCount  uint
	sendByteCount    uint
	rtpSendBuffer    []byte
	isClosed         bool
}

func NewDestination(transportMode, castMode, remoteRtpPort, remoteRtcpPort, rtpChannelId, rtcpChannelId int, conn net.Conn) *Destination {
	if transportMode == RTP_UDP {
		createUdpConn := func(port int) (*net.UDPConn, error) {
			udpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", port))
			if err != nil {
				return nil, err
			}

			conn, err := net.ListenUDP("udp", udpAddr)
			if err != nil {
				return nil, err
			}

			return conn, nil
		}

		index := strings.LastIndex(conn.RemoteAddr().String(), ":")
		remoteIP := string([]byte(conn.RemoteAddr().String())[:index])

		localRtpPort := RTP_BEGIN_SOCK_PORT

		for ; ; localRtpPort += 2 {
			rtpConn, err := createUdpConn(localRtpPort)
			if err != nil {
				continue
			}

			rtcpConn, err := createUdpConn(localRtpPort + 1)
			if err != nil {
				rtpConn.Close()

				continue
			}

			remoteRtpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", remoteIP, remoteRtpPort))
			if err != nil {
				logger.Error("rtp ResolveUDPAddr error:", err)
				return nil
			}

			remoteRtcpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", remoteIP, remoteRtcpPort))
			if err != nil {
				logger.Error("rtcp ResolveUDPAddr error:", err)
				return nil
			}

			dest := &Destination{
				TransportMode:  RTP_UDP,
				CastMode:       castMode,
				LocalRtpPort:   localRtpPort,
				LocalRtcpPort:  localRtpPort + 1,
				rtpConn:        rtpConn,
				rtcpConn:       rtcpConn,
				remoteRtpAddr:  remoteRtpAddr,
				remoteRtcpAddr: remoteRtcpAddr,
				isPlaying:      false,
				rtpSendBuffer:  make([]byte, 1920),
				isClosed:       false,
			}

			go dest.sendRTCPLoop()

			return dest
		}
	} else {
		dest := &Destination{
			TransportMode: RTP_TCP,
			CastMode:      castMode,
			RtpChannelId:  rtpChannelId,
			RtcpChannelId: rtcpChannelId,
			tcpConn:       conn,
			isPlaying:     false,
			rtpSendBuffer: make([]byte, 1920),
			isClosed:      false,
		}

		go dest.sendRTCPLoop()

		return dest
	}
}

func (dest *Destination) sendRTCPLoop() {
	sendBuffer := make([]byte, 1920)

	for !dest.isClosed {
		if dest.syncSource != 0 {
			srPack := &rtcp.RtcpPacket{
				Version:    rtcp.RTCP_VERSION,
				PacketType: rtcp.RTCP_TYPE_SR,
				SyncSource: dest.syncSource,
				Contents:   make([]io.Reader, 2),
			}

			sender := &rtcp.RtcpSender{
				RTPTimestamp: dest.currentTimestamp,
				PacketCount:  dest.sendPacketCount,
				ByteCount:    dest.sendByteCount,
			}

			srPack.Contents[0] = sender

			report := &rtcp.RTCPReceptionReport{
				SyncSource:  dest.syncSource,
				LastSeq:     dest.currentSeq,
				DelayLastSR: 20,
			}

			srPack.Contents[1] = report

			sdesPack := &rtcp.RtcpPacket{
				Version:    rtcp.RTCP_VERSION,
				PacketType: rtcp.RTCP_TYPE_SDES,
				SyncSource: dest.syncSource,
				Contents:   make([]io.Reader, 1),
			}

			sdes := &rtcp.RTCPSDES{
				SyncSource: dest.syncSource,
				Items:      make([]rtcp.RTCPSDESItem, 1),
			}

			sdes.Items[0] = rtcp.RTCPSDESItem{
				Type: rtcp.RTCP_SDES_CNAME,
				Data: "RTSPServer",
			}

			sdesPack.Contents[0] = sdes

			if dest.TransportMode == RTP_UDP {
				sendLength1, _ := srPack.Read(sendBuffer)
				sendLength2, _ := sdesPack.Read(sendBuffer[sendLength1:])
				_, err := dest.rtcpConn.WriteToUDP(sendBuffer[:sendLength1+sendLength2], dest.remoteRtcpAddr)
				if err != nil {
					return
				}
			} else {
				sendLength1, _ := srPack.Read(sendBuffer[4:])
				sendLength2, _ := sdesPack.Read(sendBuffer[4+sendLength1:])

				sendLength := sendLength1 + sendLength2

				sendBuffer[0] = '$'
				sendBuffer[1] = byte(dest.RtcpChannelId)
				sendBuffer[2] = byte((sendLength & 0xFF00) >> 8)
				sendBuffer[3] = byte(sendLength & 0xff)

				totalSendLength := 0
				for {
					writeLength, err := dest.tcpConn.Write(sendBuffer[totalSendLength : sendLength+4])
					if err != nil {
						dest.tcpConn.Close()
						return
					}

					totalSendLength += writeLength

					if totalSendLength == sendLength+4 {
						break
					}
				}
			}
		}

		time.Sleep(20 * time.Millisecond)
	}
}

func (dest *Destination) SendRtp(pack *rtp.RtpPacket) {
	dest.currentSeq = pack.SequenceNumber
	dest.currentTimestamp = pack.Timestamp
	dest.syncSource = pack.SyncSource

	var sendLength int

	if dest.TransportMode == RTP_UDP {
		sendLength, _ = pack.Read(dest.rtpSendBuffer)

		_, err := dest.rtpConn.WriteToUDP(dest.rtpSendBuffer[:sendLength], dest.remoteRtpAddr)
		if err != nil {
			return
		}
	} else {
		var err error

		sendLength, err = pack.Read(dest.rtpSendBuffer[4:])
		if err != nil {
			logger.Error("read rtp error:", err)
			return
		}

		dest.rtpSendBuffer[0] = '$'
		dest.rtpSendBuffer[1] = byte(dest.RtpChannelId)
		dest.rtpSendBuffer[2] = byte((sendLength & 0xFF00) >> 8)
		dest.rtpSendBuffer[3] = byte(sendLength & 0xff)

		totalSendLength := 0
		for {
			writeLength, err := dest.tcpConn.Write(dest.rtpSendBuffer[totalSendLength : sendLength+4])
			if err != nil {
				dest.tcpConn.Close()
				logger.Error("dest send rtp err:", err)
				break
			}

			totalSendLength += writeLength

			if totalSendLength == sendLength+4 {
				break
			}
		}
	}

	dest.sendPacketCount += 1
	dest.sendByteCount += uint(sendLength)
}

func (dest *Destination) Close() {
	dest.isClosed = true

	if dest.TransportMode == RTP_UDP {
		dest.rtpConn.Close()
		dest.rtcpConn.Close()
	} else {

	}
}
