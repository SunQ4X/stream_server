#pragma once

#ifdef _WIN32
#ifdef DEVICE_API_EXPORTS
#define DEVICE_API extern "C" _declspec(dllexport)
#else
#define DEVICE_API extern "C" _declspec(dllimport)
#endif
#else

#endif

#ifndef IN
#define IN
#endif

#ifndef OUT
#define OUT
#endif


//码流类型
typedef enum
{
	StreamTypeIntercom = 0,	//对讲流
	StreamTypeMain,		//主码流
	StreamTypeSub,		//子码流
}StreamType;

//录像类型满足条件
typedef enum
{
	RecordTypeConditionOr = 0,  /* 任意满足一种条件 */
	RecordTypeConditionAnd      /* 同时满足全部条件 */
}RecordTypeCondition;


//录像类型
typedef enum
{
	RecordType_None = 0,
	RecordType_Regular = 1 << 0, // 常规录像(1)
	RecordType_Motion = 1 << 1, // 移动侦测(2)
	RecordType_Alarm = 1 << 2, // 告警录像(3)
	RecordType_Schedule = 1 << 3, // 计划录像(8)
	RecordType_Manual = 1 << 4, // 手动录像(16)

								/* 智能侦测 */
								RecordType_Counter = 1 << 5, // 计数报警(32)
								RecordType_Wire = 1 << 6, // 跨线报警(64)
								RecordType_Region = 1 << 7, // 区域报警(128)
								RecordType_Object = 1 << 8, // 物品丢失报警(256)

								RecordType_Pre = RecordType_None, // 预录

																  /* 注意:增加一种新智能类型就修改一下AllSmartType的定义 */
																  RecordType_AllSmartType = RecordType_Counter\
																  | RecordType_Wire\
	| RecordType_Region\
	| RecordType_Object,

	RecordType_AllType = RecordType_Regular\
	| RecordType_Motion\
	| RecordType_Alarm\
	| RecordType_Schedule\
	| RecordType_Manual\
	| RecordType_AllSmartType,

	RecordType_Undefined = 1 << 9  // 未定义
}CMS_CONNECT_PARSE_RecordType;


//预置点命令字
typedef enum
{
	PresetCtrlType_Set = 1,		//设置预置点
	PresetCtrlType_Clear,		//清除预置点
	PresetCtrlType_Goto			//转到预置点
}CMS_CONNECT_PARSE_PresetCmdType;

//云台命令字
typedef enum
{
	PtzCtrlType_LightPowerOn = 1,	//灯光电源
	PtzCtrlType_WiperPowerOn,		//雨刷电源
	PtzCtrlType_FanPowerOn,			//风扇电源
	PtzCtrlType_ZoomIn,				//焦距大
	PtzCtrlType_ZoomOut,			//焦距小
	PtzCtrlType_FocusNear,			//焦点近
	PtzCtrlType_FocusFar,			//焦点远
	PtzCtrlType_IrisOpen,			//光圈开
	PtzCtrlType_IrisClose,			//光圈闭
	PtzCtrlType_TiltUp,				//上
	PtzCtrlType_TiltDown,			//下
	PtzCtrlType_PanLeft,			//左
	PtzCtrlType_PanRight,			//右
	PtzCtrlType_UpLeft,				//左上
	PtzCtrlType_UpRight,			//右上
	PtzCtrlType_DownLeft,			//左下
	PtzCtrlType_DownRight,			//右下
	PtzCtrlType_Auto				//自动
}CMS_CONNECT_PARSE_PtzCmdType;


//回放命令字
typedef enum
{
	ReplayCtrlType_IFrameOnly = 1,      //是否只弹I帧 参数 1-只弹I帧 0-弹所有帧
	ReplayCtrlType_SeekTime			//时间定位 参数 time_t(UTC)
}CMS_CONNECT_PARSE_ReplayCmdType;


typedef struct 
{
	char serial_num[64];
	int alarm_in_num;
	int alarm_out_num;
	int device_type;
	int disk_num;
	int chan_num;
	int start_chan;
	int audio_chan_num;
	int zero_chan_num;
	unsigned int surrport;
	int version;
}DeviceInfo;


typedef struct {
	int device_type;
	char serial_num[64];
	char ip[64];
	int port;
	char username[64];
	char password[64];
}DeviceSearchResult;

typedef struct {
	int preset_index;
	int dwell;
	int speed;
}CruisePoint;

typedef struct {
	int points_num;
	CruisePoint points[64];
}CruiseRoute;

typedef struct {
	int year;
	int month;
	int day;
	int hour;
	int minute;
	int second;
	int time_zone;//!时区 如:东八区 time_zone = 8 西八区 time_zone = -8  
}Time;

typedef struct {
	int ch;
	int record_type;
	Time begin_time;
	Time end_time;
}RecordParameter;

DEVICE_API void get_protocol(OUT int* protocol_id, OUT char* protocol_name, OUT int* name_length);

DEVICE_API bool start_search_device();
DEVICE_API void stop_search_device();
DEVICE_API bool get_search_device_parameter(OUT DeviceSearchResult* result);

DEVICE_API int login(IN char* ip, IN int port, IN char* username, IN char* password, OUT DeviceInfo* device_info);
DEVICE_API void logout(IN int login_handle);

//DEVICE_API bool get_device_information(IN int login_handle, OUT char* sn, OUT int* ch_num, OUT int* alarm_in_num, OUT int* alarm_out_num);

DEVICE_API int open_realtime(IN int login_handle, IN int ch, IN int stream_type);
DEVICE_API void close_realtime(IN int realtime_handle);
DEVICE_API bool get_realtime_frame(IN int realtime_handle, OUT unsigned char* data, OUT unsigned int* length);

DEVICE_API int open_intercom(IN int login_handle, IN int ch);
DEVICE_API void close_intercom(IN int intercom_handle);
DEVICE_API bool get_intercom_frame(IN int intercom_handle, OUT unsigned char* data, OUT unsigned int* length);
DEVICE_API bool send_intercom_frame(IN int intercom_handle, IN unsigned char* data, IN unsigned int length);

DEVICE_API bool ctrl_ptz(IN int login_handle, IN int ch, IN int cmd_type, IN int is_stop, IN int speed);

DEVICE_API bool set_3d(IN int login_handle, IN int ch, IN int top_x, IN int top_y, IN int bottom_x, IN int bottom_y);

DEVICE_API bool ctrl_preset(IN int login_handle, IN int ch, IN int cmd_type, IN int preset_index);

DEVICE_API bool get_cruise(IN int login_handle, IN int ch, IN int cruise_route, OUT CruiseRoute* route);
DEVICE_API bool set_cruise(IN int login_handle, IN int ch, IN int cruise_route, IN int cruise_point_index, IN CruisePoint *point);
DEVICE_API bool delete_cruise(IN int login_handle, IN int ch, IN int cruise_route, IN int cruise_point_index);
DEVICE_API bool clear_cruise(IN int login_handle, IN int ch, IN int cruise_route);
DEVICE_API bool run_cruise(IN int login_handle, IN int ch, IN int cruise_route, IN int is_run);

DEVICE_API bool ctrl_locus(IN int login_handle, IN int ch, IN int locus_index, IN int cmd_type, IN int is_stop);

//返回录像数组指针
DEVICE_API int search_record(IN int login_handle, IN Time *begin, IN Time *end, IN int iRecordType, IN int record_type_condition, IN int stream_type, IN int ch[], IN int ch_num, int *records_count);


DEVICE_API int open_replay(IN int login_handle, IN Time* begin, IN Time* end, IN int record_type, IN int record_type_condition, IN int stream_type, IN int ch[], IN int ch_num);
DEVICE_API void close_replay(IN int replay_handle);

//控制回放 CMS_CONNECT_PARSE_ReplayCmdType
DEVICE_API bool ctrl_replay(IN int replay_handle,IN int cmd_type,IN int param);

DEVICE_API bool get_replay_frame(IN int replay_handle, OUT unsigned int* channel, OUT unsigned char* data, OUT unsigned int* length);

DEVICE_API bool set_device_time(IN int login_handle, IN Time* now);

DEVICE_API int get_alarm(OUT int *login_handle, OUT int* alarm_type, OUT int* ch);

DEVICE_API bool get_config(IN int login_handle, IN int config_type, OUT unsigned char *data, IN unsigned int in_length, unsigned int* out_length);

DEVICE_API bool set_config(IN int login_handle, IN int config_type, IN unsigned char *data, IN unsigned int in_length);

DEVICE_API void release_memory(IN int bufferPtr);//释放内存