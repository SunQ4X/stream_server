package protocol

type CommonResponse struct {
	ResultCode int    `json:"resultCode"`
	ResultMsg  string `json:"resultMsg"`
}

type RtspServer struct {
	SerialNum string `json:"serialNum"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	ResultCode   int    `json:"resultCode"`
	SerialNum    string `json:"serialNum"`
	RtspUsername string `json:"rtsp_username"`
	RtspPassword string `json:rtsp_password`
}

type RtspAccount struct {
	RtspUsername string `json:"rtsp_username"`
	RtspPassword string `json:rtsp_password`
}

type GetRtspInfoResponse struct {
	ResultCode   int    `json:"resultCode"`
	RtspUsername string `json:"rtsp_username"`
	RtspPassword string `json:"rtsp_password"`
	RtspPort     int    `json:"rtsp_port"`
}

//设备信息
type DeviceInfo struct {
	Serialnum    string `json:"serialnum"`
	IP           string `json:"ip"`
	Port         int    `json:"port"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	ProtocolType int    `json:"protocolType"`
	ProtocolName string `json:"protocolName"`
	IsOnline     bool   `json:"isOnline"`
}

type ConnectInfo struct {
	Serialnum  string `json:"serialnum"`
	Prototcol  string `json:"protocol"`
	Channel    int    `json:"channel"`
	StreamType int    `json:"streamType"`
}

const (
	RtpFormatI8       = 0
	RtpFormatStandard = 1
)

//开关流请求
type OpenRealtimeRequest struct {
	SerialNum  string `json:"serialNum"`
	StreamType int    `json:"streamType"`
	Channel    int    `json:"channel"`
	Url        string `json:"url"`
	RtpFormat  int    `json:"rtp_format"`
}

type ServerStatus struct {
	TotalMemory uint64  `json:"totalMemory"`
	UsedMemory  uint64  `json:"usedMemory"`
	NetIn       uint64  `json:"netIn"`
	NetOut      uint64  `json:"netOut"`
	Cpu         float64 `json:"cpu"`
}

//录像时间
type RecordTime struct {
	Year     int32
	Month    int32
	Day      int32
	Hour     int32
	Minute   int32
	Second   int32
	TimeZone int32
}

type ChId struct {
	SerialNum string `json:"serialNum"`
	Channel   int    `json:"channel"`
}

type DeviceReplayRequest struct {
	SerialNum            string     `json:"serialNum"`
	BeginTime            RecordTime `json:"beginTime"`
	EndTime              RecordTime `json:"endTime"`
	RecordType           int        `json:"recordType"`
	StreamType           int        `json:"streamType"`
	RecordType_condition int        `json:"recordType_condition"`
	Channels             []int      `json:"channels"`
}

type StorageReplayRequest struct {
	BeginTime            RecordTime `json:"beginTime"`
	EndTime              RecordTime `json:"endTime"`
	StreamType           int        `json:"streamType"`
	RecordType           int        `json:"recordType"`
	RecordType_condition int        `json:"recordType_condition"`
	ChIds                []ChId     `json:"chIds"`
}

type OpenDeviceReplayRequest struct {
	Url           string              `json:"url"`
	ReplayRequest DeviceReplayRequest `json:"replayRequest"`
	RtpFormat     int                 `json:"rtp_format"`
}

type OpenStorageReplayRequest struct {
	Url           string               `json:"url"`
	RpcAddress    string               `json:"rpcAddress"`
	ReplayRequest StorageReplayRequest `json:"replayRequest"`
	RtpFormat     int                  `json:"rtp_format"`
}

type OpenIntercomRequest struct {
	SerialNum string `json:"serialNum"`
	Channel   int    `json:"channel"`
	Url       string `json:"url"`
}

type ClientStreamInfo struct {
	ClientAddress   string `json:"client_address"`
	RtspUrl         string `json:"rtsp_url"`
	StreamType      int    `json:"stream_type"`
	StreamParameter string `json:"stream_parameter"`
}

type GetStreamStatusResponse struct {
	ResultCode int                `json:"resultCode"`
	Streams    []ClientStreamInfo `json:"streams"`
}

//const (
//	OK             = 0
//	Failed         = 1
//	ParseFailed    = 2
//	DbFailed       = 3
//	Reject         = 4
//	NotLogin       = 5
//	NotAllow       = 6
//	DeviceNotExist = 7
//)

//返回值错误码
const (
	ResultCode_Succ            = 0
	ResultCode_Failed          = 1
	ResultCode_NoLogin         = 2
	ResultCode_HasDone         = 3
	ResultCode_ParseJsonError  = 4
	ResultCode_DbFailed        = 5
	ResultCode_InvalidParam    = 6
	ResultCode_Call3ThAPIError = 7 //调用第三方接口错误
	ResultCode_DeviceNotOnline = 8
	ResultCode_NoRtspServer    = 9 //没有转发服务器
)
