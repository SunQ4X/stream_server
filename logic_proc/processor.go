package logic_proc

import (
	"database/sql"
	"logger"
	"protocol"
	"strings"
)

const (
	DATE_INTERVAL = 0 //0表示只保存当天的数据
)

type Processor struct {
	db *sql.DB
}

var proc *Processor

func NewProcessor(db *sql.DB) *Processor {
	if nil == proc {
		proc = &Processor{
			db: db,
		}
	}
	return proc
}

func GetProcessor() *Processor {
	return proc
}

func (p *Processor) InitDatabase() error {
	//	sql := `CREATE TABLE IF NOT EXIST device (
	//  serial_num varchar(64) NOT NULL,
	//  device_ip varchar(20) NOT NULL,
	//  device_port int(11) NOT NULL,
	//  username varchar(20) DEFAULT NULL,
	//  password varchar(20) DEFAULT NULL,
	//  protocol_type int(11) DEFAULT NULL,
	//  protocol_name varchar(20) NOT NULL,
	//  is_online int(11) DEFAULT NULL,
	//  PRIMARY KEY (serial_num),
	//  UNIQUE KEY serial_num (serial_num)
	//) ENGINE=InnoDB DEFAULT CHARSET=utf8;
	//`

	//	_, err := p.db.Exec(sql)
	//	if err != nil {
	//		return err
	//	}

	return nil
}

//登录并获取设备信息，只允许一个中心服务器连接
func (p *Processor) Login(request protocol.LoginRequest) (response protocol.LoginResponse) {
	row := p.db.QueryRow("SELECT password FROM user WHERE username=?", request.Username)
	var password string
	err := row.Scan(&password)
	if err != nil {
		logger.Error("login db failed:", err)
		response.ResultCode = protocol.ResultCode_DbFailed
		return
	}

	if password != request.Password {
		logger.Error("login  failed:username or password is wrong")
		response.ResultCode = protocol.ResultCode_Failed
		return
	}

	row = p.db.QueryRow("SELECT rtsp_username,rtsp_password FROM rtsp_account")
	err = row.Scan(&response.RtspUsername, &response.RtspPassword)
	if err != nil {
		logger.Error(err)
		response.ResultCode = protocol.ResultCode_DbFailed
		return
	}

	return
}

func (p *Processor) SetLoginAccount(request protocol.LoginRequest) protocol.CommonResponse {
	_, err := p.db.Exec("UPDATE user SET username=?,password=?", request.Username, request.Password)
	var response protocol.CommonResponse
	if err != nil {
		response.ResultCode = protocol.ResultCode_DbFailed
	}
	return response

}

//获取rtsp账号
func (p *Processor) GetRtspAccount() (username, password string, err error) {
	row := p.db.QueryRow("SELECT rtsp_username,rtsp_password FROM rtsp_account")
	err = row.Scan(&username, &password)
	if err != nil {
		logger.Error(err)
	}
	return
}

func (p *Processor) SetRtspAccount(request protocol.RtspAccount) (response protocol.CommonResponse) {
	_, err := p.db.Exec("UPDATE rtsp_account SET rtsp_username=?,rtsp_password=?", request.RtspUsername, request.RtspPassword)
	if err != nil {
		response.ResultCode = protocol.ResultCode_DbFailed
		logger.Error(err)
	}
	return
}

func (p *Processor) SetDeviceParam(request protocol.DeviceInfo) protocol.CommonResponse {
	var response protocol.CommonResponse
	_, err := p.db.Exec("REPLACE INTO device(serial_num,device_ip,device_port,username,password,protocol_type,protocol_name,is_online) VALUES(?,?,?,?,?,?,?,?)",
		request.Serialnum, request.IP, request.Port, request.Username, request.Password, request.ProtocolType, strings.ToUpper(request.ProtocolName), request.IsOnline)
	if err != nil {
		logger.Error(err)
		response.ResultCode = protocol.ResultCode_DbFailed
	}
	return response
}

func (p *Processor) GetDeviceParam(serialNum string) (response protocol.DeviceInfo, err error) {
	row := p.db.QueryRow("SELECT device_ip,device_port,username,password,protocol_type,protocol_name FROM device WHERE serial_num=?", serialNum)
	err = row.Scan(&response.IP, &response.Port, &response.Username, &response.Password,
		&response.ProtocolType, &response.ProtocolName)
	return
}

func (p *Processor) DelDevice(serialNum string) error {
	_, err := p.db.Exec("DELETE FROM device WHERE serial_num=?", serialNum)
	return err
}

func (p *Processor) AddStream(url string, streamType int, paramJson string) error {
	_, err := p.db.Exec("REPLACE INTO stream(stream_url, stream_type, param_json) VALUES(?,?,?)", url, streamType, paramJson)
	if err != nil {
		logger.Error("add stream failed", url, paramJson, err)
		return err
	}

	return nil
}

func (p *Processor) GetAllStreamParam() ([]StreamInformation, error) {
	rows, err := p.db.Query("SELECT stream_url,stream_type,param_json FROM stream")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	informations := make([]StreamInformation, 0, 10)
	for rows.Next() {
		var info StreamInformation
		if err := rows.Scan(&info.Url, &info.StreamType, &info.ParamJson); err == nil {
			informations = append(informations, info)
		}
	}

	return informations, nil
}

func (p *Processor) AddClient(address, streamName string) error {
	_, err := p.db.Exec("INSERT INTO client(client_address, stream_url) VALUES(?,?)", address, streamName)
	return err
}

func (p *Processor) RemoveClient(address string) error {
	_, err := p.db.Exec("DELETE FROM client WHERE client_address=?", address)
	return err
}

func (p *Processor) GetClients() ([]protocol.ClientStreamInfo, error) {
	rows, err := p.db.Query("SELECT client.client_address, client.stream_url, stream.stream_type, stream.param_json FROM client INNER JOIN stream ON client.stream_url=stream.stream_url")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	clients := make([]protocol.ClientStreamInfo, 0, 10)
	for rows.Next() {
		var info protocol.ClientStreamInfo
		if err := rows.Scan(&info.ClientAddress, &info.RtspUrl, &info.StreamType, &info.StreamParameter); err == nil {
			clients = append(clients, info)
		}
	}

	return clients, nil
}

func (p *Processor) ClearClients() error {
	_, err := p.db.Exec("DELETE FROM client")
	return err

}

type StreamInformation struct {
	Url        string
	StreamType int
	ParamJson  string
}
