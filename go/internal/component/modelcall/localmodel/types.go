// Package localmodel 提供了本地 AI 模型管理的核心功能
// 包括服务器管理、任务调度、配置管理等功能
package localmodel

// ServerStatus 表示服务器的运行状态
type ServerStatus string

const (
	ServerStopped  ServerStatus = "stopped"  // 服务器已停止
	ServerStarting ServerStatus = "starting" // 服务器正在启动
	ServerRunning  ServerStatus = "running"  // 服务器正在运行
	ServerStopping ServerStatus = "stopping" // 服务器正在停止
	ServerError    ServerStatus = "error"    // 服务器出现错误
)

// ServerType 表示服务器的类型
type ServerType string

const (
	ServerLocal    ServerType = "local"    // 本地服务器
	ServerLocalDir ServerType = "localDir" // 本地目录服务器
	ServerCloud    ServerType = "cloud"    // 云端服务器
)

// ServerFunction 表示服务器支持的功能类型
type ServerFunction string

const (
	FunctionVideoGen     ServerFunction = "videoGen"     // 视频生成功能
	FunctionSoundTts     ServerFunction = "soundTts"     // 语音合成功能
	FunctionSoundClone   ServerFunction = "soundClone"   // 语音克隆功能
	FunctionAsr          ServerFunction = "asr"          // 语音识别功能
	FunctionLive         ServerFunction = "live"         // 直播功能
	FunctionImageGen     ServerFunction = "imageGen"     // 图像生成功能
	FunctionImageEdit    ServerFunction = "imageEdit"    // 图像编辑功能
	FunctionImageUpscale ServerFunction = "imageUpscale" // 图像放大功能
)

// ServerSetting 表示服务器的设置项
type ServerSetting struct {
	Name        string `json:"name"`        // 设置名称
	Type        string `json:"type"`        // 设置类型
	Title       string `json:"title"`       // 设置标题
	Default     string `json:"default"`     // 默认值
	Placeholder string `json:"placeholder"` // 占位符
	Options     []struct {
		Value string `json:"value"` // 选项值
		Label string `json:"label"` // 选项标签
	} `json:"options,omitempty"` // 选项列表
}

// ServerConfig 表示服务器的配置信息
type ServerConfig struct {
	Name              string           `json:"name"`              // 服务器名称
	Version           string           `json:"version"`           // 服务器版本
	Title             string           `json:"title"`             // 服务器标题
	Description       string           `json:"description"`       // 服务器描述
	DeviceDescription string           `json:"deviceDescription"` // 设备描述
	PlatformName      string           `json:"platformName"`      // 平台名称
	PlatformArch      string           `json:"platformArch"`      // 平台架构
	ServerRequire     string           `json:"serverRequire"`     // 服务器要求
	Entry             string           `json:"entry"`             // 入口点
	Functions         []ServerFunction `json:"functions"`         // 支持的功能列表
	Settings          []ServerSetting  `json:"settings"`          // 设置列表
	EasyServer        *struct {
		Entry     string   `json:"entry"`     // EasyServer 入口点
		EntryArgs []string `json:"entryArgs"` // EasyServer 入口参数
		Envs      []string `json:"envs"`      // 环境变量
		Content   string   `json:"content"`   // 内容
	} `json:"easyServer,omitempty"` // EasyServer 特定配置
}

// ServerRecord 表示服务器记录
type ServerRecord struct {
	Key       string                 `json:"key"`       // 服务器唯一标识
	Name      string                 `json:"name"`      // 服务器名称
	Title     string                 `json:"title"`     // 服务器标题
	Version   string                 `json:"version"`   // 服务器版本
	Type      ServerType             `json:"type"`      // 服务器类型
	Functions []ServerFunction       `json:"functions"` // 支持的功能列表
	LocalPath string                 `json:"localPath"` // 本地路径
	AutoStart bool                   `json:"autoStart"` // 是否自动启动
	Settings  []ServerSetting        `json:"settings"`  // 设置列表
	Setting   map[string]interface{} `json:"setting"`   // 设置值
	Status    ServerStatus           `json:"status"`    // 服务器状态
	Runtime   *ServerRuntime         `json:"runtime"`   // 运行时信息
	Config    *ServerConfig          `json:"config"`    // 服务器配置
}

// ServerRuntime 表示服务器的运行时信息
type ServerRuntime struct {
	Status           ServerStatus `json:"status"`                     // 运行状态
	AutoStartStatus  ServerStatus `json:"autoStartStatus"`            // 自动启动状态
	LogFile          string       `json:"logFile"`                    // 日志文件路径
	StartTimestampMS int64        `json:"startTimestampMS,omitempty"` // 启动时间戳（毫秒）
	EventChannelName string       `json:"eventChannelName,omitempty"` // 事件通道名称
	StartTime        int64        `json:"startTime,omitempty"`        // 启动时间
}

// ServerInfo 表示服务器的信息
type ServerInfo struct {
	LocalPath        string                 `json:"localPath"`        // 本地路径
	Name             string                 `json:"name"`             // 服务器名称
	Version          string                 `json:"version"`          // 服务器版本
	Setting          map[string]interface{} `json:"setting"`          // 设置
	LogFile          string                 `json:"logFile"`          // 日志文件路径
	EventChannelName string                 `json:"eventChannelName"` // 事件通道名称
	Config           ServerRecord           `json:"config"`           // 服务器配置
}

// ServerFunctionDataType 表示服务器功能的数据类型
type ServerFunctionDataType struct {
	ID          string                 `json:"id"`                    // 任务ID
	Result      map[string]interface{} `json:"result"`                // 结果数据
	Param       map[string]interface{} `json:"param,omitempty"`       // 参数
	Text        string                 `json:"text,omitempty"`        // 文本内容
	Video       string                 `json:"video,omitempty"`       // 视频文件路径
	Audio       string                 `json:"audio,omitempty"`       // 音频文件路径
	PromptAudio string                 `json:"promptAudio,omitempty"` // 提示音频
	PromptText  string                 `json:"promptText,omitempty"`  // 提示文本
}

// LauncherResultType 表示启动器的结果类型
type LauncherResultType struct {
	Result  map[string]interface{} `json:"result"`  // 结果数据
	EndTime *int64                 `json:"endTime"` // 结束时间
}

// SoundTtsModelConfig 表示语音合成模型的配置
type SoundTtsModelConfig struct {
	Type  string                 `json:"type"`  // 模型类型
	Param map[string]interface{} `json:"param"` // 参数
	Text  string                 `json:"text"`  // 文本内容
}

// SoundCloneModelConfig 表示语音克隆模型的配置
type SoundCloneModelConfig struct {
	Type        string                 `json:"type"`        // 模型类型
	Param       map[string]interface{} `json:"param"`       // 参数
	Text        string                 `json:"text"`        // 文本内容
	PromptAudio string                 `json:"promptAudio"` // 提示音频
	PromptText  string                 `json:"promptText"`  // 提示文本
}

// VideoGenModelConfig 表示视频生成模型的配置
type VideoGenModelConfig struct {
	Type  string                 `json:"type"`  // 模型类型
	Param map[string]interface{} `json:"param"` // 参数
	Video string                 `json:"video"` // 视频文件路径
	Audio string                 `json:"audio"` // 音频文件路径
}

// AsrModelConfig 表示语音识别模型的配置
type AsrModelConfig struct {
	Audio string                 `json:"audio"` // 音频文件路径
	Param map[string]interface{} `json:"param"` // 参数
}

// TaskStatus 表示任务的状态
type TaskStatus string

const (
	TaskQueue   TaskStatus = "queue"   // 队列中
	TaskWait    TaskStatus = "wait"    // 等待中
	TaskRunning TaskStatus = "running" // 运行中
	TaskSuccess TaskStatus = "success" // 成功
	TaskFail    TaskStatus = "fail"    // 失败
	TaskPause   TaskStatus = "pause"   // 暂停
)

// TaskRecord 表示任务记录
type TaskRecord struct {
	ID            int64                  `json:"id,omitempty"`     // 任务ID
	Biz           string                 `json:"biz"`              // 业务类型
	Type          string                 `json:"type"`             // 任务类型
	Title         string                 `json:"title"`            // 任务标题
	Status        TaskStatus             `json:"status"`           // 任务状态
	StatusMsg     string                 `json:"statusMsg"`        // 状态消息
	StartTime     int64                  `json:"startTime"`        // 开始时间
	EndTime       *int64                 `json:"endTime"`          // 结束时间
	ServerName    string                 `json:"serverName"`       // 服务器名称
	ServerTitle   string                 `json:"serverTitle"`      // 服务器标题
	ServerVersion string                 `json:"serverVersion"`    // 服务器版本
	Param         map[string]interface{} `json:"param"`            // 参数
	JobResult     map[string]interface{} `json:"jobResult"`        // 作业结果
	ModelConfig   map[string]interface{} `json:"modelConfig"`      // 模型配置
	Result        map[string]interface{} `json:"result"`           // 结果
	Runtime       map[string]interface{} `json:"runtime"`          // 运行时信息
}

// TaskResult 表示任务的执行结果
type TaskResult struct {
	Code int         `json:"code"` // 状态码
	Msg  string      `json:"msg"`  // 消息
	Data interface{} `json:"data"` // 数据
}
