package utility

import (
	"encoding/json"
	"io/ioutil"
)

const (
	OPTIONS_CONFIG_FILE_PATH = "./options.cfg"
	RTSP_DEFAULT             = ":554"
	HTTP_DEFAULT             = ":80"
	MYSQL_HOST_ADDR_DEFAULT  = "127.0.0.1:3306"
	MYSQL_DBNAME_DEFAULT     = "cms_rtsp"
	MYSQL_USERNAME_DEFAULT   = "root"
	MYSQL_PASSWORD_DEFAULT   = ""
)

type Options struct {
	RTSPAddress   string `json:rtsp_address`
	HTTPAddress   string `json:http_address`
	MysqlAddress  string `json:"mysql_address"`
	MysqlDbName   string `json:"mysql_db_name"`
	MysqlUsername string `json:"mysql_username"`
	MysqlPassword string `json:"mysql_password"`
}

var opts *Options

func init() {
	opts = &Options{
		RTSPAddress:   RTSP_DEFAULT,
		HTTPAddress:   HTTP_DEFAULT,
		MysqlAddress:  MYSQL_HOST_ADDR_DEFAULT,
		MysqlDbName:   MYSQL_DBNAME_DEFAULT,
		MysqlUsername: MYSQL_USERNAME_DEFAULT,
		MysqlPassword: MYSQL_PASSWORD_DEFAULT,
	}

	data, err := ioutil.ReadFile(OPTIONS_CONFIG_FILE_PATH)
	if err != nil {
		//文件不存在则创建文件并写入默认配置
		data, err = json.Marshal(opts)
		if err != nil {
			return
		}

		err = ioutil.WriteFile(OPTIONS_CONFIG_FILE_PATH, data, 0666)
	}
	//将data转成结构体opts
	json.Unmarshal(data, opts)
}

func GetOptions() *Options {
	return opts
}
