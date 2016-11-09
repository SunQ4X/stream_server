package rtsp

type Controler interface {
	Control(req *Request) *Response
}
