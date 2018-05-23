package stream_server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"logger"
	"logic_proc"
	"protocol"
	"rtsp_server"
	"stream"
	"sync"
	"utility"
	"web_api"
	//"github.com/garyburd/redigo/redis"

	_ "github.com/go-sql-driver/mysql"
	//	_ "github.com/mattn/go-sqlite3"
)

type StreamServer struct {
	wrap sync.WaitGroup
}

func NewStreamServer() *StreamServer {
	return &StreamServer{}
}

func (s *StreamServer) Wrap(cb func()) {
	s.wrap.Add(1)
	go func() {
		cb()
		s.wrap.Done()
	}()
}

func (s *StreamServer) Run() {
	dataSource := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8", utility.GetOptions().MysqlUsername, utility.GetOptions().MysqlPassword, utility.GetOptions().MysqlAddress, utility.GetOptions().MysqlDbName)
	db, err := sql.Open("mysql", dataSource)
	//db, err := sql.Open("sqlite3", "cms_rtsp.db")
	if err != nil {
		logger.Error("Open database failed, err:", err)
		return
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		logger.Error("ping database failed, err:", err)
		return
	}

	logic_proc.NewProcessor(db)

	err = logic_proc.GetProcessor().InitDatabase()
	if err != nil {
		logger.Error("Init database failed, err:", err)
		return
	}

	err = logic_proc.GetProcessor().ClearClients()
	if err != nil {
		logger.Error("Clear clients failed, err:", err)
		return
	}

	stream.LoadDeviceLibraries()

	informations, err := logic_proc.GetProcessor().GetAllStreamParam()
	if err != nil {
		logger.Error("Recover stream failed, err:", err)
		return
	}

	for _, info := range informations {
		switch info.StreamType {
		case stream.STREAM_TYPE_REALTIME_I8, stream.STREAM_TYPE_REALTIME_STANDARD:
			var param protocol.OpenRealtimeRequest
			if err := json.Unmarshal([]byte(info.ParamJson), &param); err == nil {
				stream.AddStream(param.Url, param)
				logger.Debug("add stream:", param.Url)
			}
		case stream.STREAM_TYPE_INTERCOM_I8:
			var param protocol.OpenIntercomRequest
			if err := json.Unmarshal([]byte(info.ParamJson), &param); err == nil {
				stream.AddStream(param.Url, param)
				logger.Debug("add stream:", param.Url)
			}
		case stream.STREAM_TYPE_DEVICE_REPLAY_I8, stream.STREAM_TYPE_DEVICE_REPLAY_STANDARD:
			var param protocol.OpenDeviceReplayRequest
			if err := json.Unmarshal([]byte(info.ParamJson), &param); err == nil {
				stream.AddStream(param.Url, param)
				logger.Debug("add stream:", param.Url)
			}
		case stream.STREAM_TYPE_STORAGE_REPLAY_I8, stream.STREAM_TYPE_STORAGE_REPLAY_STANDARD:
			var param protocol.OpenStorageReplayRequest
			if err := json.Unmarshal([]byte(info.ParamJson), &param); err == nil {
				stream.AddStream(param.Url, param)
				logger.Debug("add stream:", param.Url)
			}
		}
	}

	rtspServer, err := rtsp_server.NewRtspServer()
	if err != nil {
		logger.Error(err.Error())
		return
	}

	apiServer, err := web_api.NewServer()
	if err != nil {
		logger.Error(err.Error())
		return
	}

	s.Wrap(func() {
		rtspServer.Run()
	})

	s.Wrap(func() {
		apiServer.Run()
	})

	s.wrap.Wait()
}
