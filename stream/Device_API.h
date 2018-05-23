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


//��������
typedef enum
{
	StreamTypeIntercom = 0,	//�Խ���
	StreamTypeMain,		//������
	StreamTypeSub,		//������
}StreamType;

//¼��������������
typedef enum
{
	RecordTypeConditionOr = 0,  /* ��������һ������ */
	RecordTypeConditionAnd      /* ͬʱ����ȫ������ */
}RecordTypeCondition;


//¼������
typedef enum
{
	RecordType_None = 0,
	RecordType_Regular = 1 << 0, // ����¼��(1)
	RecordType_Motion = 1 << 1, // �ƶ����(2)
	RecordType_Alarm = 1 << 2, // �澯¼��(3)
	RecordType_Schedule = 1 << 3, // �ƻ�¼��(8)
	RecordType_Manual = 1 << 4, // �ֶ�¼��(16)

								/* ������� */
								RecordType_Counter = 1 << 5, // ��������(32)
								RecordType_Wire = 1 << 6, // ���߱���(64)
								RecordType_Region = 1 << 7, // ���򱨾�(128)
								RecordType_Object = 1 << 8, // ��Ʒ��ʧ����(256)

								RecordType_Pre = RecordType_None, // Ԥ¼

																  /* ע��:����һ�����������;��޸�һ��AllSmartType�Ķ��� */
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

	RecordType_Undefined = 1 << 9  // δ����
}CMS_CONNECT_PARSE_RecordType;


//Ԥ�õ�������
typedef enum
{
	PresetCtrlType_Set = 1,		//����Ԥ�õ�
	PresetCtrlType_Clear,		//���Ԥ�õ�
	PresetCtrlType_Goto			//ת��Ԥ�õ�
}CMS_CONNECT_PARSE_PresetCmdType;

//��̨������
typedef enum
{
	PtzCtrlType_LightPowerOn = 1,	//�ƹ��Դ
	PtzCtrlType_WiperPowerOn,		//��ˢ��Դ
	PtzCtrlType_FanPowerOn,			//���ȵ�Դ
	PtzCtrlType_ZoomIn,				//�����
	PtzCtrlType_ZoomOut,			//����С
	PtzCtrlType_FocusNear,			//�����
	PtzCtrlType_FocusFar,			//����Զ
	PtzCtrlType_IrisOpen,			//��Ȧ��
	PtzCtrlType_IrisClose,			//��Ȧ��
	PtzCtrlType_TiltUp,				//��
	PtzCtrlType_TiltDown,			//��
	PtzCtrlType_PanLeft,			//��
	PtzCtrlType_PanRight,			//��
	PtzCtrlType_UpLeft,				//����
	PtzCtrlType_UpRight,			//����
	PtzCtrlType_DownLeft,			//����
	PtzCtrlType_DownRight,			//����
	PtzCtrlType_Auto				//�Զ�
}CMS_CONNECT_PARSE_PtzCmdType;


//�ط�������
typedef enum
{
	ReplayCtrlType_IFrameOnly = 1,      //�Ƿ�ֻ��I֡ ���� 1-ֻ��I֡ 0-������֡
	ReplayCtrlType_SeekTime			//ʱ�䶨λ ���� time_t(UTC)
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
	int time_zone;//!ʱ�� ��:������ time_zone = 8 ������ time_zone = -8  
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

//����¼������ָ��
DEVICE_API int search_record(IN int login_handle, IN Time *begin, IN Time *end, IN int iRecordType, IN int record_type_condition, IN int stream_type, IN int ch[], IN int ch_num, int *records_count);


DEVICE_API int open_replay(IN int login_handle, IN Time* begin, IN Time* end, IN int record_type, IN int record_type_condition, IN int stream_type, IN int ch[], IN int ch_num);
DEVICE_API void close_replay(IN int replay_handle);

//���ƻط� CMS_CONNECT_PARSE_ReplayCmdType
DEVICE_API bool ctrl_replay(IN int replay_handle,IN int cmd_type,IN int param);

DEVICE_API bool get_replay_frame(IN int replay_handle, OUT unsigned int* channel, OUT unsigned char* data, OUT unsigned int* length);

DEVICE_API bool set_device_time(IN int login_handle, IN Time* now);

DEVICE_API int get_alarm(OUT int *login_handle, OUT int* alarm_type, OUT int* ch);

DEVICE_API bool get_config(IN int login_handle, IN int config_type, OUT unsigned char *data, IN unsigned int in_length, unsigned int* out_length);

DEVICE_API bool set_config(IN int login_handle, IN int config_type, IN unsigned char *data, IN unsigned int in_length);

DEVICE_API void release_memory(IN int bufferPtr);//�ͷ��ڴ�