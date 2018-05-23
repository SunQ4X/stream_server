package rtsp

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	RTSP_VERSION = "RTSP/1.0"
)

const (
	// Client to server for presentation and stream objects; recommended
	DESCRIBE = "DESCRIBE"
	// Bidirectional for client and stream objects; optional
	ANNOUNCE = "ANNOUNCE"
	// Bidirectional for client and stream objects; optional
	GET_PARAMETER = "GET_PARAMETER"
	// Bidirectional for client and stream objects; required for Client to server, optional for server to client
	OPTIONS = "OPTIONS"
	// Client to server for presentation and stream objects; recommended
	PAUSE = "PAUSE"
	// Client to server for presentation and stream objects; required
	PLAY = "PLAY"
	// Client to server for presentation and stream objects; optional
	RECORD = "RECORD"
	// Server to client for presentation and stream objects; optional
	REDIRECT = "REDIRECT"
	// Client to server for stream objects; required
	SETUP = "SETUP"
	// Bidirectional for presentation and stream objects; optional
	SET_PARAMETER = "SET_PARAMETER"
	// Client to server for presentation and stream objects; required
	TEARDOWN = "TEARDOWN"
	//自定义扩展
	ANTSCOMB_ADDCH = "ANTSCOMB_ADDCH"
)

const (
	// all requests
	Continue = 100

	// all requests
	OK = 200
	// RECORD
	Created = 201
	// RECORD
	LowOnStorageSpace = 250

	// all requests
	MultipleChoices = 300
	// all requests
	MovedPermanently = 301
	// all requests
	MovedTemporarily = 302
	// all requests
	SeeOther = 303
	// all requests
	UseProxy = 305

	// all requests
	BadRequest = 400
	// all requests
	Unauthorized = 401
	// all requests
	PaymentRequired = 402
	// all requests
	Forbidden = 403
	// all requests
	NotFound = 404
	// all requests
	MethodNotAllowed = 405
	// all requests
	NotAcceptable = 406
	// all requests
	ProxyAuthenticationRequired = 407
	// all requests
	RequestTimeout = 408
	// all requests
	Gone = 410
	// all requests
	LengthRequired = 411
	// DESCRIBE, SETUP
	PreconditionFailed = 412
	// all requests
	RequestEntityTooLarge = 413
	// all requests
	RequestURITooLong = 414
	// all requests
	UnsupportedMediaType = 415
	// SETUP
	Invalidparameter = 451
	// SETUP
	IllegalConferenceIdentifier = 452
	// SETUP
	NotEnoughBandwidth = 453
	// all requests
	SessionNotFound = 454
	// all requests
	MethodNotValidInThisState = 455
	// all requests
	HeaderFieldNotValid = 456
	// PLAY
	InvalidRange = 457
	// SET_PARAMETER
	ParameterIsReadOnly = 458
	// all requests
	AggregateOperationNotAllowed = 459
	// all requests
	OnlyAggregateOperationAllowed = 460
	// all requests
	UnsupportedTransport = 461
	// all requests
	DestinationUnreachable = 462

	// all requests
	InternalServerError = 500
	// all requests
	NotImplemented = 501
	// all requests
	BadGateway = 502
	// all requests
	ServiceUnavailable = 503
	// all requests
	GatewayTimeout = 504
	// all requests
	RTSPVersionNotSupported = 505
	// all requests
	OptionNotsupport = 551
)

// RTSP请求的格式
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// | 方法  | <空格>  | URL | <空格>  | 版本  | <回车换行>  |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// | 头部字段名  | : | <空格>  |      值     | <回车换行>  |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                   ......                              |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// | 头部字段名  | : | <空格>  |      值     | <回车换行>  |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// | <回车换行>                                            |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |  实体内容                                             |
// |  （通常不用）                                         |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
type Request struct {
	Method  string
	URL     string
	Version string
	Header  map[string]string
	Body    string
}

func NewRequest(method, url, cSeq, body string) *Request {
	req := &Request{
		Method:  method,
		URL:     url,
		Version: RTSP_VERSION,
		Header:  make(map[string]string),
		Body:    body,
	}

	req.Header["CSeq"] = cSeq

	return req
}

func (r *Request) String() string {
	str := fmt.Sprintf("%s %s %s\r\n", r.Method, r.URL, r.Version)
	for key, value := range r.Header {
		str += fmt.Sprintf("%s: %s\r\n", key, value)
	}
	str += "\r\n"
	str += r.Body
	return str
}

func ReadRequest(str string) (req *Request, err error) {
	defer func() {
		if re := recover(); re != nil {
			req = nil
			err = errors.New("run time panic")
		}
	}()

	req = &Request{
		Header: make(map[string]string),
	}

	context := strings.SplitN(str, "\r\n\r\n", 2)
	header := context[0]
	body := context[1]

	parts := strings.SplitN(header, "\r\n", 2)
	dest := parts[0]
	prop := parts[1]

	parts = strings.SplitN(dest, " ", 3)
	req.Method = parts[0]
	req.URL = parts[1]
	req.Version = parts[2]

	pairs := strings.Split(prop, "\r\n")
	for _, pair := range pairs {
		parts = strings.SplitN(pair, ": ", 2)
		key := parts[0]
		value := parts[1]
		req.Header[key] = value
	}

	req.Body = string(body)

	return req, nil
}

// RTSP响应的格式
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// | 版本  | <空格>  | 状态码 | <空格>  | 状态描述  | <回车换行> |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// | 头部字段名  | : | <空格>  |      值           | <回车换行>  |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                           ......                            |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// | 头部字段名  | : | <空格>  |      值           | <回车换行>  |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// | <回车换行>                                                  |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |  实体内容                                                   |
// |  （有些响应不用）                                           |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
type Response struct {
	Version    string
	StatusCode int
	Status     string
	Header     map[string]string
	Body       string
}

func NewResponse(statusCode int, status, cSeq, body string) *Response {
	res := &Response{
		Version:    RTSP_VERSION,
		StatusCode: statusCode,
		Status:     status,
		Header:     make(map[string]string),
		Body:       body,
	}

	res.Header["CSeq"] = cSeq

	return res
}

func (r *Response) String() string {
	str := fmt.Sprintf("%s %d %s\r\n", r.Version, r.StatusCode, r.Status)
	for key, value := range r.Header {
		str += fmt.Sprintf("%s: %s\r\n", key, value)
	}
	str += "\r\n"
	str += r.Body
	return str
}

func ReadResponse(str string) (res *Response, err error) {
	defer func() {
		if re := recover(); re != nil {
			fmt.Println("run time panic:", re)
			res = nil
			err = errors.New("run time panic")
		}
	}()

	res = &Response{
		Header: make(map[string]string),
	}

	context := strings.SplitN(str, "\r\n\r\n", 2)
	header := context[0]
	body := context[1]

	parts := strings.SplitN(header, "\r\n", 2)
	status := parts[0]
	prop := parts[1]

	parts = strings.SplitN(status, " ", 3)
	res.Version = parts[0]
	if res.StatusCode, err = strconv.Atoi(parts[1]); err != nil {
		return nil, err
	}
	res.Status = parts[2]

	pairs := strings.Split(prop, "\r\n")
	for _, pair := range pairs {
		parts = strings.SplitN(pair, ": ", 2)
		key := parts[0]
		value := parts[1]
		res.Header[key] = value
	}

	res.Body = string(body)

	return res, nil
}
