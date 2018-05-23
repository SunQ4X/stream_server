package stream

/*
#cgo linux CFLAGS: -DLINUX=1 -fPIC
#cgo LDFLAGS: -L . -lIntelligence
#include "libIntelligence.h"
*/
import "C"

import (
	//	"encoding/json"
	"logger"
	"logic_proc"
	"os"
	"path/filepath"
	"protocol"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"
	"utility"

	//	"github.com/garyburd/redigo/redis"
)

const (
	DEVICE_PROTOCOL_LIB_DIR = "./device_libs"
)

var (
	deviceCache     = utility.NewCache()
	deviceProtocols = make(map[string]*DeviceProtocol)
	loginMutex      = &sync.Mutex{}
)

type DeviceProtocol struct {
	login             uintptr
	logout            uintptr
	openRealtime      uintptr
	closeRealtime     uintptr
	getRealtimeFrame  uintptr
	openReplay        uintptr
	closeReplay       uintptr
	getReplayFrame    uintptr
	ctrlReplay        uintptr
	openIntercom      uintptr
	closeIntercom     uintptr
	getIntercomFrame  uintptr
	sendIntercomFrame uintptr
}

func LoadDeviceLibraries() {
	logger.Info("设备库目录:", DEVICE_PROTOCOL_LIB_DIR)
	filepath.Walk(DEVICE_PROTOCOL_LIB_DIR, func(path string, info os.FileInfo, err error) error {
		defer func() {
			if re := recover(); re != nil {
				logger.Error("加载设备库异常:", re)
			}
		}()

		if info != nil && !info.IsDir() {
			h, err := syscall.LoadLibrary(path)
			if err != nil {
				logger.Error("加载设备库文件", path, "失败:", err.Error())
				return nil
			}

			getProtocol, err := syscall.GetProcAddress(h, "get_protocol")
			if err != nil {
				logger.Error("加载设备库文件", path, "函数[get_protocol]失败", err.Error())
				return nil
			}

			login, err := syscall.GetProcAddress(h, "login")
			if err != nil {
				logger.Error("加载设备库文件", path, "函数[login]失败", err.Error())
				return nil
			}

			logout, err := syscall.GetProcAddress(h, "logout")
			if err != nil {
				logger.Error("加载设备库文件", path, "函数[logout]失败", err.Error())
				return nil
			}

			openRealtime, err := syscall.GetProcAddress(h, "open_realtime")
			if err != nil {
				logger.Error("加载设备库文件", path, "函数[open_realtime]失败", err.Error())
				return nil
			}

			closeRealtime, err := syscall.GetProcAddress(h, "close_realtime")
			if err != nil {
				logger.Error("加载设备库文件", path, "函数[close_realtime]失败", err.Error())
				return nil
			}

			getRealtimeFrame, err := syscall.GetProcAddress(h, "get_realtime_frame")
			if err != nil {
				logger.Error("加载设备库文件", path, "函数[get_realtime_frame]失败", err.Error())
				return nil
			}

			openReplay, err := syscall.GetProcAddress(h, "open_replay")
			if err != nil {
				logger.Error("加载设备库文件", path, "函数[open_replay]失败", err.Error())
				return nil
			}

			closeReplay, err := syscall.GetProcAddress(h, "close_replay")
			if err != nil {
				logger.Error("加载设备库文件", path, "函数[close_replay]失败", err.Error())
				return nil
			}

			getReplayFrame, err := syscall.GetProcAddress(h, "get_replay_frame")
			if err != nil {
				logger.Error("加载设备库文件", path, "函数[get_realtime_frame]失败", err.Error())
				return nil
			}

			ctrlReplay, err := syscall.GetProcAddress(h, "ctrl_replay")
			if err != nil {
				logger.Error("加载设备库文件", path, "函数[ctrl_replay]失败", err.Error())
				return nil
			}

			openIntercom, err := syscall.GetProcAddress(h, "open_intercom")
			if err != nil {
				logger.Error("加载设备库文件", path, "函数[open_intercom]失败", err.Error())
				return nil
			}

			closeIntercom, err := syscall.GetProcAddress(h, "close_intercom")
			if err != nil {
				logger.Error("加载设备库文件", path, "函数[close_intercom]失败", err.Error())
				return nil
			}

			getIntercomFrame, err := syscall.GetProcAddress(h, "get_intercom_frame")
			if err != nil {
				logger.Error("加载设备库文件", path, "函数[get_intercom_frame]失败", err.Error())
				return nil
			}

			sendIntercomFrame, err := syscall.GetProcAddress(h, "send_intercom_frame")
			if err != nil {
				logger.Error("加载设备库文件", path, "函数[send_intercom_frame]失败", err.Error())
				return nil
			}

			protocolId := 0
			protocolNameBuffer := make([]byte, 256)
			protocolNameLength := 0

			_, _, errno := syscall.Syscall(getProtocol,
				3,
				uintptr(unsafe.Pointer(&protocolId)),
				uintptr(unsafe.Pointer(&protocolNameBuffer[0])),
				uintptr(unsafe.Pointer(&protocolNameLength)))

			if errno != 0 {
				logger.Error("加载设备库文件", path, "执行函数[getProtocol]失败", errno.Error())
				return nil
			}

			deviceProtocol := &DeviceProtocol{
				login:             login,
				logout:            logout,
				openRealtime:      openRealtime,
				closeRealtime:     closeRealtime,
				getRealtimeFrame:  getRealtimeFrame,
				openReplay:        openReplay,
				closeReplay:       closeReplay,
				getReplayFrame:    getReplayFrame,
				ctrlReplay:        ctrlReplay,
				openIntercom:      openIntercom,
				closeIntercom:     closeIntercom,
				getIntercomFrame:  getIntercomFrame,
				sendIntercomFrame: sendIntercomFrame,
			}

			protocolName := string(protocolNameBuffer[:protocolNameLength])

			deviceProtocols[strings.ToUpper(protocolName)] = deviceProtocol

			logger.Info("加载设备协议", protocolName, "成功")
		}

		return nil
	})

	logger.Info("设备库加载完毕")
}

type Device struct {
	deviceProtocol *DeviceProtocol
	serialNum      string
	loginHandle    int
	*utility.ReferenceCounter
}

func GetDevice(serialNum string) (device *Device) {
	defer func() {
		if device != nil {
			device.AddReference()
		}
	}()

	if element, ok := deviceCache.Lookup(serialNum); ok {
		device = element.(*Device)
		return
	}

	loginMutex.Lock()
	defer loginMutex.Unlock()

	if element, ok := deviceCache.Lookup(serialNum); ok {
		device = element.(*Device)
		return
	}

	//获取设备连接参数,连接并登录设备
	deviceParam, err := logic_proc.GetProcessor().GetDeviceParam(serialNum)
	if err != nil {
		logger.Error("获取设备信息出错", serialNum, err.Error())
		return
	}

	deviceProtocol, ok := deviceProtocols[deviceParam.ProtocolName]
	if !ok {
		logger.Error("没有设备协议", deviceParam.ProtocolName)
		return
	}

	logger.Debug("设备登录")

	var loginHandle uintptr
	var errno syscall.Errno

	if deviceParam.Password == "" {
		loginHandle, _, errno = syscall.Syscall6(deviceProtocol.login,
			5,
			uintptr(unsafe.Pointer(&[]byte(deviceParam.IP)[0])),
			uintptr(deviceParam.Port),
			uintptr(unsafe.Pointer(&[]byte(deviceParam.Username)[0])),
			0,
			0,
			0)
	} else {
		loginHandle, _, errno = syscall.Syscall6(deviceProtocol.login,
			5,
			uintptr(unsafe.Pointer(&[]byte(deviceParam.IP)[0])),
			uintptr(deviceParam.Port),
			uintptr(unsafe.Pointer(&[]byte(deviceParam.Username)[0])),
			uintptr(unsafe.Pointer(&[]byte(deviceParam.Password)[0])),
			0,
			0)
	}

	if errno != 0 {
		logger.Error("设备", serialNum, "登录出错:", errno.Error())
		return
	}

	if loginHandle <= 0 {
		logger.Info("设备", serialNum, "登录失败")
		return
	}

	device = &Device{
		deviceProtocol:   deviceProtocol,
		serialNum:        serialNum,
		loginHandle:      int(loginHandle),
		ReferenceCounter: utility.NewReferenceCounter(serialNum, time.Minute*1),
	}

	deviceCache.Add(serialNum, device)

	return
}

func (device *Device) Erase() {
	logger.Debug("设备注销")

	_, _, errno := syscall.Syscall(device.deviceProtocol.logout,
		1,
		uintptr(device.loginHandle),
		0,
		0)

	if errno != 0 {
		logger.Error("设备", device.serialNum, "注销出错:", errno.Error())
	}
}

func (device *Device) OpenRealtime(ch, stream_type int) int {
	logger.Info("打开实时流")
	handle, _, errno := syscall.Syscall(device.deviceProtocol.openRealtime,
		3,
		uintptr(device.loginHandle),
		uintptr(ch),
		uintptr(stream_type))

	if errno != 0 {
		logger.Error("设备", device.serialNum, "打开实时流出错:", errno.Error())

		return 0
	}

	if handle <= 0 {
		logger.Error("设备", device.serialNum, "打开实时流失败")

		return 0
	}

	return int(handle)
}

func (device *Device) CloseRealtime(handle int) {
	logger.Info("关闭实时流")
	_, _, errno := syscall.Syscall(device.deviceProtocol.closeRealtime,
		1,
		uintptr(handle),
		0,
		0)

	if errno != 0 {
		logger.Error("设备", device.serialNum, "关闭实时流出错:", errno.Error())
	}
}

func (device *Device) GetRealtimeData(handle int, data []byte) (int, error) {
	var length int

	_, _, errno := syscall.Syscall(device.deviceProtocol.getRealtimeFrame,
		3,
		uintptr(handle),
		uintptr(unsafe.Pointer(&data[0])),
		uintptr(unsafe.Pointer(&length)))

	if errno != 0 {
		//		logger.Error("设备", device.serialNum, "获取实时流数据出错:", errno.Error())
		return 0, errno
	}

	return length, nil
}

func (device *Device) OpenReplay(request protocol.DeviceReplayRequest) int {
	logger.Info("打开历史流")

	replayHandle, _, errno := syscall.Syscall9(device.deviceProtocol.openReplay,
		8,
		uintptr(device.loginHandle),
		uintptr(unsafe.Pointer(&request.BeginTime)),
		uintptr(unsafe.Pointer(&request.EndTime)),
		uintptr(request.RecordType),
		uintptr(request.RecordType_condition), //查询条件
		uintptr(request.StreamType),           //码流类型
		uintptr(unsafe.Pointer(&request.Channels[0])),
		uintptr(len(request.Channels)),
		0)

	if errno != 0 {
		logger.Error("设备", device.serialNum, "打开历史流出错:", errno.Error())
		return 0
	}

	if replayHandle <= 0 {
		logger.Error("设备", device.serialNum, "打开历史流失败")
		return 0
	}

	logger.Info("open Repaly handle:", replayHandle)

	return int(replayHandle)
}

func (device *Device) CloseReplay(replayHandle int) {
	logger.Info("关闭历史流")

	_, _, errno := syscall.Syscall(device.deviceProtocol.closeReplay,
		1,
		uintptr(replayHandle),
		0,
		0)

	if errno != 0 {
		logger.Error("设备", device.serialNum, "关闭回放出错:", errno.Error())
	}
}

//获取回放帧
func (device *Device) GetReplayData(handle int, data []byte) (int, int, error) {
	var length, ch int

	_, _, errno := syscall.Syscall6(device.deviceProtocol.getReplayFrame,
		4,
		uintptr(handle),
		uintptr(unsafe.Pointer(&ch)),
		uintptr(unsafe.Pointer(&data[0])),
		uintptr(unsafe.Pointer(&length)),
		0,
		0)

	if errno != 0 {
		logger.Error("设备", device.serialNum, "获取回放流数据出错:", errno.Error())
		return 0, 0, nil
	}

	//logger.Debug("设备获取历史流", "通道:", ch)

	return ch, length, nil
}

func (device *Device) SeekReplayTime(handle, year, month, day, hour, minute, second, timeZone int) error {
	seekTime := time.Date(year, time.Month(month), day, hour, minute, second, 0, time.FixedZone("", timeZone))

	_, _, errno := syscall.Syscall(device.deviceProtocol.ctrlReplay,
		3,
		uintptr(handle),
		uintptr(2),
		uintptr(seekTime.Unix()))

	if errno != 0 {
		logger.Error("设备", device.serialNum, "定位回放时间出错:", errno.Error())
		return errno
	}

	logger.Debug("定位回放时间", seekTime)
	return nil
}

func (device *Device) OpenIntercom(ch int) int {
	logger.Info("打开对讲流")

	handle, _, errno := syscall.Syscall(device.deviceProtocol.openIntercom,
		2,
		uintptr(device.loginHandle),
		uintptr(ch),
		0)

	if errno != 0 {
		logger.Error("设备", device.serialNum, "打开对讲流出错:", errno.Error())

		return 0
	}

	if handle <= 0 {
		logger.Error("设备", device.serialNum, "打开对讲流失败")

		return 0
	}

	return int(handle)
}

func (device *Device) CloseIntercom(handle int) {
	logger.Info("关闭对讲流")

	_, _, errno := syscall.Syscall(device.deviceProtocol.closeIntercom,
		1,
		uintptr(handle),
		0,
		0)

	if errno != 0 {
		logger.Error("设备", device.serialNum, "关闭对讲流出错:", errno.Error())
	}
}

func (device *Device) GetIntercomData(handle int, data []byte) (int, error) {
	var length int
	_, _, errno := syscall.Syscall(device.deviceProtocol.getIntercomFrame,
		3,
		uintptr(handle),
		uintptr(unsafe.Pointer(&data[0])),
		uintptr(unsafe.Pointer(&length)))

	if errno != 0 {
		logger.Error("设备", device.serialNum, "获取对讲流数据出错:", errno.Error())
		return 0, errno
	}

	return length, nil
}

func (device *Device) SendIntercomData(handle int, data []byte) error {
	_, _, errno := syscall.Syscall(device.deviceProtocol.sendIntercomFrame,
		3,
		uintptr(handle),
		uintptr(unsafe.Pointer(&data[0])),
		uintptr(len(data)))

	if errno != 0 {
		logger.Error("设备", device.serialNum, "发送对讲流数据出错:", errno.Error())
		return errno
	}

	return nil
}
