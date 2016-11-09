package stream_server

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/stream_server/rtsp"
)

type RtspControler struct {
	server *StreamServer
}

func (r *RtspControler) Control(req *rtsp.Request) *rtsp.Response {
	cSeq := req.Header.Get("CSeq")

	reqUrl, err := url.ParseRequestURI(req.URL)
	if err != nil {
		return rtsp.NewResponse(rtsp.BadGateway, "Url parse error", cSeq, "")
	}

	switch req.Method {
	case rtsp.OPTIONS:
		resp := rtsp.NewResponse(rtsp.OK, "OK", cSeq, "")
		options := strings.Join([]string{rtsp.OPTIONS, rtsp.DESCRIBE, rtsp.SETUP, rtsp.PLAY, rtsp.TEARDOWN}, ", ")
		resp.Header["Public"] = []string{options}
		return resp
	case rtsp.DESCRIBE:
		path := reqUrl.Path
		fmt.Println("request path:", path)
	case rtsp.SETUP:

	case rtsp.PLAY:

	case rtsp.ANNOUNCE:

	case rtsp.RECORD:

	case rtsp.TEARDOWN:

	default:
		return rtsp.NewResponse(rtsp.MethodNotAllowed, "Option Not Support", cSeq, "")
	}

	return rtsp.NewResponse(rtsp.OK, "OK", cSeq, "")
}
