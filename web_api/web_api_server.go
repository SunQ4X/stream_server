package web_api

import (
	"encoding/json"
	//	"fmt"
	"io/ioutil"
	"logger"
	"logic_proc"
	"strconv"
	//	"media_session"
	"net"
	"net/http"
	"protocol"
	"stream"
	"strings"
	"time"
	"utility"
	"webserver"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	//"github.com/shirou/gopsutil/net"
	"net/http/pprof"
)

var SERVER_MAC_ID = ""

func init() {
	interfaces, err := net.Interfaces()
	if err != nil {
		panic("Get server MAC failed, err:" + err.Error())
	}

	SERVER_MAC_ID = strings.Replace(interfaces[0].HardwareAddr.String(), ":", "", -1) + "aa" //转发服务器用aa表示
	logger.Info("Get server MAC:", SERVER_MAC_ID)
}

type Server struct {
	*webserver.WebServer
	processor *logic_proc.Processor
}

func byteString(p []byte) string {
	for i := 0; i < len(p); i++ {
		if p[i] == 0 {
			return string(p[0:i])
		}
	}
	return string(p)
}

func NewServer() (*Server, error) {
	webServer, err := webserver.NewWebServer(utility.GetOptions().HTTPAddress)
	if err != nil {
		return nil, err
	}

	processor := logic_proc.GetProcessor()

	server := &Server{
		WebServer: webServer,
		processor: processor,
	}

	server.NotFoundHandler(server.getResource)
	server.Handle("GET", "/", server.getAPI) //获取相关api接口

	server.Handle("POST", "/cms3/loginServer", server.loginServer)  //中心服务器登录转发
	server.Handle("POST", "/cms3/setAccount", server.setAccount)    //设置转发服务器账号
	server.Handle("GET", "/cms3/info/rtsp", server.getRtspInfo)     //获取rtsp信息
	server.Handle("POST", "/cms3/info/rtsp", server.setRtspAccount) //设置rtsp信息

	server.Handle("GET", "/cms3/serverStatus", server.getServerStatus) //获取服务器状态

	server.Handle("POST", "/cms3/setDeviceParam", server.setDeviceLoginParam) //设置设备登录参数
	server.Handle("POST", "/cms3/getDeviceParam", server.getDeviceLoginParam) //获取
	//server.Handle("POST", "/cms3/delDeviceParam", server.delDevice)           //删除设备

	server.Handle("POST", "/cms3/realtime", server.openRealTime)            //打开预览
	server.Handle("POST", "/cms3/device/replay", server.openDeviceReplay)   //打开设备回放
	server.Handle("POST", "/cms3/storage/replay", server.openStorageReplay) //打开存储服务器回放
	server.Handle("POST", "/cms3/intercom", server.openIntercom)            //打开对讲

	server.Handle("GET", "/cms3/heartBeat", server.heartBeat) //心跳

	server.Handle("GET", "/cms3/status/stream", server.getStreamStatus) //获取码流状态

	return server, nil
}

func (s *Server) getResource(w http.ResponseWriter, r *http.Request) {
	if methodUpper := strings.ToUpper(r.Method); methodUpper != "GET" {
		http.NotFound(w, r)
		return
	}

	if strings.HasPrefix(r.URL.Path, "/static/") {
		http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))).ServeHTTP(w, r)
	} else if strings.HasPrefix(r.URL.Path, "/view/") /*&& strings.HasSuffix(r.URL.Path, ".html") */ {
		http.StripPrefix("/view/", http.FileServer(http.Dir("./view"))).ServeHTTP(w, r)
	} else if strings.HasPrefix(r.URL.Path, "/debug/pprof/") {
		pprof.Index(w, r)
	} else {
		http.NotFound(w, r)
	}
}

func (s *Server) getAPI(w http.ResponseWriter, r *http.Request) {
	//logger.Info("request")
	http.Redirect(w, r, "/view/index.html", http.StatusFound)
}

/**-
*@api{GET} /cms3/heartBeat 心跳
*@apiGroup Users
 * @apiVersion 1.1.1
 * @apiDescription 心跳
 * @apiSuccess (200) {String} msg 信息
 * @apiSuccess (200) {int} code 0 代表无错误 1代表有错误
 * @apiSuccessExample {json} 返回样例:
 *                {"code":"0","msg":""}
*/
func (s *Server) heartBeat(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	var response protocol.CommonResponse
	defer func() {
		resp, _ := json.Marshal(response)
		w.Write(resp)
	}()
}

/**-
*@api{POST} /cms3/loginServer 登录
*@apiGroup Users
 * @apiVersion 1.1.1
 * @apiDescription 登录转发服务器
 * @apiParam {String} username 用户名称
 * @apiParam {String} password 密码
 * @apiParamExample {json} 请求样例：
 *                {
				"username":"admin",
				"password":"123456"
			}
 * @apiSuccess (200) {int} resultCode 0 代表无错误 1代表有错误
 * @apiSuccess (200) {String} rtsp_username rtsp开流账号
 * @apiSuccess (200) {String} rtsp_password rtsp开流密码
 * @apiSuccess (200) {String} serialNum 服务器唯一标识
 * @apiSuccessExample {json} 返回样例:
 *                {
				"resultCode":"0",
				"rtsp_username":"admin",
				"rtsp_password":"123456",
				"serialNum":"12346abc"
			}
*/
func (s *Server) loginServer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	var response protocol.LoginResponse
	var request protocol.LoginRequest
	defer func() {
		resp, _ := json.Marshal(response)
		w.Write(resp)
	}()

	body, _ := ioutil.ReadAll(r.Body)

	err := json.Unmarshal(body, &request)
	if err != nil {
		logger.Error(err)
		response.ResultCode = protocol.ResultCode_DbFailed
		return
	}

	response = s.processor.Login(request)
	if 0 == response.ResultCode {
		response.SerialNum = SERVER_MAC_ID
	}
}

/**
 *@api {POST} /cms3/setAccount 设置转发服务器账号
 *@apiGroup Users
 * @apiVersion 1.1.1
 * @apiDescription 用于设置账号密码
 * @apiParam {String} username 用户名称
 * @apiParam {String} password 密码
 * @apiParamExample {json} 请求样例：
 *                {"username":"admin","password":"123456"}
 * @apiSuccess (200) {String} resultMsg 信息
 * @apiSuccess (200) {int} resultCode 0 代表无错误 1代表有错误
 * @apiSuccessExample {json} 返回样例:
 *                {"resultCode":"0","resultMsg":"设置成功"}
 */
func (s *Server) setAccount(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	var response protocol.CommonResponse
	var request protocol.LoginRequest
	defer func() {
		resp, _ := json.Marshal(response)
		w.Write(resp)
	}()

	body, _ := ioutil.ReadAll(r.Body)

	err := json.Unmarshal(body, &request)
	if err != nil {
		response.ResultCode = protocol.ResultCode_ParseJsonError
		response.ResultMsg = "request error:" + err.Error()
		return
	}
	response = s.processor.SetLoginAccount(request)
}

/**
 * @api {GET} /cms3/info/rtsp 获取rtsp连接账号
 * @apiGroup Users
 * @apiVersion 1.1.1
 * @apiDescription 用于获取账号密码端口
 * @apiSuccess (200) {String} rtsp_username  用户名
 * @apiSuccess (200) {String} rtsp_password  密码
 * @apiSuccess (200) {String} rtsp_port  端口
 * @apiSuccessExample {json} 返回样例:
 *               {"resultCode":"0","rtsp_username":"admin","rtsp_password":"123456","rtsp_port":554}
 */
func (s *Server) getRtspInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	//	jsonstr, err := s.processor.GetRtspAccount()
	//	if err != nil {

	//	} else {
	//		w.Write([]byte(jsonstr))
	//	}
	var response protocol.GetRtspInfoResponse
	defer func() {
		resp, _ := json.Marshal(response)
		w.Write(resp)
	}()
	username, password, err := s.processor.GetRtspAccount()
	if err != nil {
		response.ResultCode = protocol.ResultCode_DbFailed
	}

	logger.Debug(username, password)

	port := strings.SplitN(utility.GetOptions().RTSPAddress, ":", 2)[1]

	response.RtspUsername = username
	response.RtspPassword = password
	response.RtspPort, _ = strconv.Atoi(port)
}

/**
 * @api {POST} /cms3/info/rtsp 设置rtsp连接账号信息
 * @apiGroup Users
 * @apiVersion 1.1.1
 * @apiDescription 用于设置rtsp连接的账号信息
 * @apiParam {String} username 用户名称
 * @apiParam {String} password 密码
 * @apiParamExample {json} 请求样例：
 *                {"username":"admin","password":"123456"}
 * @apiSuccess (200) {String} msg 信息
 * @apiSuccess (200) {int} code 0 代表无错误 1代表有错误
 * @apiSuccessExample {json} 返回样例:
 *                {"code":"0","msg":"设置成功"}
 */
func (s *Server) setRtspAccount(w http.ResponseWriter, r *http.Request) {
	var request protocol.RtspAccount
	var response protocol.CommonResponse

	defer func() {
		resp, _ := json.Marshal(response)
		w.Write(resp)
	}()

	body, _ := ioutil.ReadAll(r.Body)

	err := json.Unmarshal(body, &request)
	if err != nil {
		response.ResultCode = protocol.ResultCode_ParseJsonError
		response.ResultMsg = "request error:" + err.Error()
		return
	}

	response = s.processor.SetRtspAccount(request)
}

/**
 * @api {GET} /cms3/serverStatus 获取服务器运行状态
 * @apiGroup Server
 * @apiVersion 1.1.1
 * @apiDescription 用于获取服务器运行状态
 * @apiSuccessExample {json} 返回样例:
 *                {"totalMemory":5123466,"usedMemory":4216564,"cpu":54.25,"netIn":45614646,"netOut":164687316}
 */
//获取服务器运行状态
func (s *Server) getServerStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	var response protocol.ServerStatus

	v, _ := mem.SwapMemory()
	response.TotalMemory = v.Total
	response.UsedMemory = v.Used
	response.Cpu = v.UsedPercent
	response.NetIn = v.Sin
	response.NetOut = v.Sout

	//获取cpu的使用率，false是total，true是每个核
	c, _ := cpu.Percent(time.Second, false)
	response.Cpu = c[0]

	//	n, _ := net.IOCounters(false)
	//	response.NetIn = n[0].BytesRecv
	//	response.NetOut = n[0].BytesSent

	resp, _ := json.Marshal(response)

	w.Write(resp)
}

/**
 * @api {POST} /cms3/setDeviceParam 设置设备登录信息
 * @apiGroup Device
 * @apiVersion 1.1.1
 * @apiDescription 用于获取服务器运行状态
 * @apiParamExample {json} 请求样例：
*			{
			   "serialnum":"02614581a214a345f458",
			   "protocolName":"ONVIF",
			   "protocolType":4,
			   "username":"admin",
			   "password":"123456",
			   "ip":"10.0.0.33",
			   "port":8000
			}
 * @apiSuccessExample {json} 返回样例:
 *                {"code":"0","msg":"设置成功"}
*/
//设置设备登录参数
func (s *Server) setDeviceLoginParam(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	var response protocol.CommonResponse

	defer func() {
		resp, _ := json.Marshal(response)
		w.Write(resp)
	}()

	body, _ := ioutil.ReadAll(r.Body)
	var request protocol.DeviceInfo

	if err := json.Unmarshal(body, &request); err != nil {
		logger.Error("Unmarshal error", err.Error())
		response.ResultCode = protocol.ResultCode_ParseJsonError
		response.ResultMsg = "parse json request to struct failed"
		return
	}

	if "" == request.Serialnum || "" == request.IP || 0 == request.Port || "" == request.ProtocolName {
		response.ResultCode = protocol.ResultCode_InvalidParam
		return
	}

	//s.processor.
	response = s.processor.SetDeviceParam(request)
	//response.ErrorCode = protocol.OK
	//response.Msg = "No Error"
}

/**
 * @api {POST} /cms3/getDeviceParam 获取设备的登录信息(暂停使用)
 * @apiGroup Device
 * @apiVersion 1.1.1
 * @apiDescription 用于获取设备信息
 * @apiParamExample {json} 请求样例：
*			{
			   "serialnum":"02614581a214a345f458"
			}
 * @apiSuccessExample {json} 返回样例:
 *               		{
				   "serialnum":"02614581a214a345f458",
				   "protocol":"ONVIF",
				   "username":"admin",
				   "password":"123456",
				   "ip":"10.0.0.33",
				   "port":8000
				}
*/
//获取设备登录参数
func (s *Server) getDeviceLoginParam(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

}

/**
 * @api {POST} /cms3/realtime 开实时流
 * @apiGroup Device
 * @apiVersion 1.1.1
 * @apiDescription 用于打开流
 * @apiParamExample {json} 请求样例：
*			[
				{
					"serialnum":"02614581a214a345f458",
					"streamType":0,
					"channel":1,
					"url":"10.0.0.106:554/ch01"
				},
				{
					"serialnum":"02614581a214a345f458",
					"streamType":0,
					"channel":2,
					"url":"10.0.0.106:554/ch02"
				}
			]
 * @apiSuccessExample {json} 返回样例:
 *               		{
					"resultCode":0,
					"resultMsg":"指令发送 成功"
				}
*/
//开流
func (s *Server) openRealTime(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	body, _ := ioutil.ReadAll(r.Body)
	var response protocol.CommonResponse

	defer func() {
		resp, _ := json.Marshal(response)
		w.Write(resp)
	}()

	//解析请求
	requests := make([]protocol.OpenRealtimeRequest, 1)
	if err := json.Unmarshal(body, &requests); err != nil {
		logger.Error("Unmarshal error", err.Error())
		response.ResultCode = protocol.ResultCode_ParseJsonError
		response.ResultMsg = "parse json request to struct failed"
		return
	}

	for i := 0; i < len(requests); i++ {
		b, err := json.Marshal(requests[i])
		if err != nil {
			continue
		}

		streamType := stream.STREAM_TYPE_REALTIME_I8
		if requests[i].RtpFormat == protocol.RtpFormatStandard {
			streamType = stream.STREAM_TYPE_REALTIME_STANDARD
		}

		err = s.processor.AddStream(requests[i].Url, streamType, string(b))
		if err != nil {
			continue
		}

		stream.AddStream(requests[i].Url, requests[i])
	}
	logger.Info("openstream end.")

	response.ResultCode = protocol.ResultCode_Succ
}

/**
 * @api {POST} /cms3/openIntercom 对讲
 * @apiGroup Device
 * @apiVersion 1.1.1
 * @apiDescription 用于打开对讲流
 * @apiParamExample {json} 请求样例：
*				{
					"serialnum":"02614581a214a345f458",
					"channel":1,
					"url":"?type=intercom&sn=02614581a214a345f458&ch=1"
				}
 * @apiSuccessExample {json} 返回样例:
 *               		{
					"resultCode":0,
					"resultMsg":"指令发送 成功"
				}
*/
func (s *Server) openIntercom(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	body, _ := ioutil.ReadAll(r.Body)
	var response protocol.CommonResponse

	defer func() {
		resp, _ := json.Marshal(response)
		w.Write(resp)
	}()

	var request protocol.OpenIntercomRequest
	if err := json.Unmarshal(body, &request); err != nil {
		logger.Error("parse error", body)
		response.ResultCode = protocol.ResultCode_ParseJsonError
		return
	}

	err := s.processor.AddStream(request.Url, stream.STREAM_TYPE_INTERCOM_I8, string(body))
	if err != nil {
		response.ResultCode = protocol.ResultCode_DbFailed
		return
	}

	stream.AddStream(request.Url, request)

	response.ResultCode = protocol.ResultCode_Succ
}

/**
 * @api {POST} /cms3/device/replay 打开回放
 * @apiGroup Device
 * @apiVersion 1.1.1
 * @apiDescription 用于打开回放流
 * @apiParamExample {json} 请求样例：
*			{
				"url":"xxxxx",
				"replayRequest":{
					"beginTime":{
									"year":2017,
									"month":11,
									"day":23,
									"hour":0,
									"minute":0,
									"second":0,
									"timeZone":0
								},
					"endTime":{
									"year":2017,
									"month":11,
									"day":23,
									"hour":0,
									"minute":0,
									"second":0,
									"timeZone":0
								},
					"recordType":1,
					"recordType_condition":0,
					"streamType":0,
					"serialNum":"029a012326e6e8f1b4dc",
					"channels":[
								1,
								2
							]
				}
			}
 * @apiSuccessExample {json} 返回样例:
 *               		{
					"resultCode":0,
					"resultMsg":"指令发送 成功"
				}
*/

func (s *Server) openDeviceReplay(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	body, _ := ioutil.ReadAll(r.Body)
	var response protocol.CommonResponse

	defer func() {
		resp, _ := json.Marshal(response)
		w.Write(resp)
	}()

	var request protocol.OpenDeviceReplayRequest
	if err := json.Unmarshal(body, &request); err != nil {
		logger.Error("parse error", body)
		response.ResultCode = protocol.ResultCode_ParseJsonError
		return
	}

	streamType := stream.STREAM_TYPE_REALTIME_I8
	if request.RtpFormat == protocol.RtpFormatStandard {
		streamType = stream.STREAM_TYPE_REALTIME_STANDARD
	}

	err := s.processor.AddStream(request.Url, streamType, string(body))
	if err != nil {
		response.ResultCode = protocol.ResultCode_DbFailed
		return
	}

	stream.AddStream(request.Url, request)

	response.ResultCode = protocol.ResultCode_Succ
}

/**
 * @api {POST} /cms3/storage/replay 打开存储服务器的回放
 * @apiGroup Device
 * @apiVersion 1.1.1
 * @apiDescription 用于打开回放流
 * @apiParamExample {json} 请求样例：
*			{
				url:"xxxxx",
				replayRequest:{
					"beginTime":{
									"year":2017,
									"month":11,
									"day":23,
									"hour":0,
									"minute":0,
									"second":0,
									"timeZone":0
								},
								"endTime":{
									"year":2017,
									"month":11,
									"day":23,
									"hour":0,
									"minute":0,
									"second":0,
									"timeZone":0
								},
					"streamType":1,
					"recordType":1,
					"recordType_condition":0,
					"chId":[
								{
									"serialNum":"0123456789",
									"channel":1
								}
							]
				}
			}
 * @apiSuccessExample {json} 返回样例:
 *               		{
					"resultCode":0,
					"resultMsg":"指令发送 成功"
				}
*/
func (s *Server) openStorageReplay(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	body, _ := ioutil.ReadAll(r.Body)
	var response protocol.CommonResponse

	defer func() {
		resp, _ := json.Marshal(response)
		w.Write(resp)
	}()

	var request protocol.OpenStorageReplayRequest
	if err := json.Unmarshal(body, &request); err != nil {
		logger.Error("parse error", body)
		response.ResultCode = protocol.ResultCode_ParseJsonError
		return
	}

	streamType := stream.STREAM_TYPE_REALTIME_I8
	if request.RtpFormat == protocol.RtpFormatStandard {
		streamType = stream.STREAM_TYPE_REALTIME_STANDARD
	}

	err := s.processor.AddStream(request.Url, streamType, string(body))
	if err != nil {
		response.ResultCode = protocol.ResultCode_DbFailed
		return
	}

	stream.AddStream(request.Url, request)

	response.ResultCode = protocol.ResultCode_Succ
}

func (s *Server) getStreamStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	var response protocol.GetStreamStatusResponse
	defer func() {
		resp, _ := json.Marshal(response)
		w.Write(resp)
	}()

	clients, err := s.processor.GetClients()
	if err != nil {
		response.ResultCode = protocol.ResultCode_DbFailed
		return
	}

	response.ResultCode = protocol.ResultCode_Succ
	response.Streams = clients
}
