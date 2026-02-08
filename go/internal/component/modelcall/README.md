# AIGC Panel Go SDK

一个用于管理和运行本地 AI 模型服务器的 Go 语言 SDK，支持语音合成、语音克隆、视频生成、语音识别等多种 AI 功能。

## 功能特性

- 🎵 **语音合成 (TTS)** - 将文本转换为语音
- 🎤 **语音克隆** - 基于提示音频克隆语音
- 🎬 **视频生成** - 生成视频内容
- 🎧 **语音识别 (ASR)** - 将语音转换为文本
- 🚀 **简单易用** - 提供简洁的 API 接口
- 🔧 **配置灵活** - 支持 JSON 配置文件
- 📊 **状态管理** - 完整的服务器状态监控

## 安装

### 前提条件

- Go 1.19 或更高版本

### 使用 go get 安装

```bash
go get github.com/zk3151643/aigcpanel-go
```

### 在项目中导入

```go
import (
    "github.com/zk3151643/aigcpanel-go"
    "github.com/zk3151643/aigcpanel-go/easyserver"
    "github.com/zk3151643/aigcpanel-go/localmodel"
)
```

## 快速开始

### 1. 准备配置文件

创建一个 JSON 配置文件 `config.json`：

```json
{
    "name": "my-ai-model",
    "version": "1.0.0",
    "title": "我的AI模型",
    "description": "AI模型服务器",
    "platformName": "darwin",
    "platformArch": "arm64",
    "entry": "server.py",
    "functions": ["soundTts", "soundClone", "videoGen", "asr"],
    "settings": [
        {
            "name": "port",
            "type": "number",
            "title": "端口",
            "default": "8080"
        }
    ]
}
```

### 2. 基本使用示例

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/zk3151643/aigcpanel-go"
    "github.com/zk3151643/aigcpanel-go/easyserver"
)

func main() {
    // 加载配置
    config, err := aigcpanel.LoadConfigFromJSON("config.json")
    if err != nil {
        log.Fatal("加载配置失败:", err)
    }
    
    // 创建服务器实例
    server := easyserver.NewEasyServer("/path/to/model", config)
    
    // 启动服务器
    if err := server.Start(); err != nil {
        log.Fatal("启动服务器失败:", err)
    }
    defer server.Stop()
    
    // 使用语音合成功能
    result, err := server.SoundTts(easyserver.ServerFunctionDataType{
        ID:   "tts-001",
        Text: "你好，世界！",
        Param: map[string]interface{}{
            "speaker": "中文女",
            "speed":   1.0,
        },
    })
    
    if err != nil {
        log.Fatal("语音合成失败:", err)
    }
    
    fmt.Printf("任务结果: %+v\n", result)
}
```

## 核心组件

### 1. EasyServer - 简易服务器管理

`easyserver` 包提供了简单易用的服务器管理功能：

#### 主要类型

```go
// EasyServer 服务器实例
type EasyServer struct {
    // 内部字段...
}

// ServerConfig 服务器配置
type ServerConfig struct {
    Name         string           `json:"name"`         // 服务器名称
    Version      string           `json:"version"`      // 版本
    Title        string           `json:"title"`        // 标题
    Description  string           `json:"description"`  // 描述
    Entry        string           `json:"entry"`        // 入口点
    Functions    []ServerFunction `json:"functions"`    // 支持的功能
    Settings     []ServerSetting  `json:"settings"`     // 设置项
}

// ServerFunctionDataType 功能调用数据
type ServerFunctionDataType struct {
    ID          string                 `json:"id"`          // 任务ID
    Text        string                 `json:"text"`        // 文本内容
    Audio       string                 `json:"audio"`       // 音频文件
    Video       string                 `json:"video"`       // 视频文件
    PromptAudio string                 `json:"promptAudio"` // 提示音频
    PromptText  string                 `json:"promptText"`  // 提示文本
    Param       map[string]interface{} `json:"param"`       // 参数
}
```

#### 主要方法

```go
// 创建新的服务器实例
func NewEasyServer(modelPath string, config *ServerConfig) *EasyServer

// 启动服务器
func (s *EasyServer) Start() error

// 停止服务器
func (s *EasyServer) Stop() error

// 获取服务器状态
func (s *EasyServer) GetStatus() ServerStatus

// AI 功能调用方法
func (s *EasyServer) SoundTts(data ServerFunctionDataType) (*TaskResult, error)
func (s *EasyServer) SoundClone(data ServerFunctionDataType) (*TaskResult, error)
func (s *EasyServer) VideoGen(data ServerFunctionDataType) (*TaskResult, error)
func (s *EasyServer) Asr(data ServerFunctionDataType) (*TaskResult, error)
```

### 2. LocalModel - 本地模型管理

`localmodel` 包提供了更高级的本地模型管理功能：

#### 主要类型

```go
// ServerManager 服务器管理器
type ServerManager struct {
    // 内部字段...
}

// ServerRecord 服务器记录
type ServerRecord struct {
    Key       string                 `json:"key"`       // 唯一标识
    Name      string                 `json:"name"`      // 名称
    Title     string                 `json:"title"`     // 标题
    Version   string                 `json:"version"`   // 版本
    Type      ServerType             `json:"type"`      // 类型
    Functions []ServerFunction       `json:"functions"` // 功能列表
    LocalPath string                 `json:"localPath"` // 本地路径
    Status    ServerStatus           `json:"status"`    // 状态
    Settings  []ServerSetting        `json:"settings"`  // 设置
    Setting   map[string]interface{} `json:"setting"`   // 设置值
}

// TaskRecord 任务记录
type TaskRecord struct {
    ID          int64                  `json:"id"`          // 任务ID
    Type        string                 `json:"type"`        // 任务类型
    Title       string                 `json:"title"`       // 标题
    Status      TaskStatus             `json:"status"`      // 状态
    StartTime   int64                  `json:"startTime"`   // 开始时间
    EndTime     *int64                 `json:"endTime"`     // 结束时间
    Param       map[string]interface{} `json:"param"`       // 参数
    Result      map[string]interface{} `json:"result"`      // 结果
}
```

### 3. 配置加载器

```go
// 从 JSON 文件加载配置
func LoadConfigFromJSON(configPath string) (*easyserver.ServerConfig, error)
```

## 详细使用示例

### 语音合成 (TTS)

```go
// 语音合成示例
func textToSpeech() {
    config, _ := aigcpanel.LoadConfigFromJSON("config.json")
    server := easyserver.NewEasyServer("/path/to/model", config)
    
    server.Start()
    defer server.Stop()
    
    result, err := server.SoundTts(easyserver.ServerFunctionDataType{
        ID:   "tts-001",
        Text: "欢迎使用AIGC Panel Go SDK",
        Param: map[string]interface{}{
            "speaker":    "中文女",
            "speed":      1.2,
            "pitch":      1.0,
            "volume":     0.8,
        },
    })
    
    if err != nil {
        log.Printf("语音合成失败: %v", err)
        return
    }
    
    if result.Code == 0 {
        fmt.Printf("语音合成成功，音频文件: %v\n", result.Data)
    }
}
```

### 语音克隆

```go
// 语音克隆示例
func voiceCloning() {
    config, _ := aigcpanel.LoadConfigFromJSON("config.json")
    server := easyserver.NewEasyServer("/path/to/model", config)
    
    server.Start()
    defer server.Stop()
    
    result, err := server.SoundClone(easyserver.ServerFunctionDataType{
        ID:          "clone-001",
        Text:        "这是克隆的声音",
        PromptAudio: "/path/to/reference.wav",
        PromptText:  "参考音频的文本内容",
        Param: map[string]interface{}{
            "temperature": 0.7,
            "top_p":       0.9,
        },
    })
    
    if err != nil {
        log.Printf("语音克隆失败: %v", err)
        return
    }
    
    fmt.Printf("语音克隆结果: %+v\n", result)
}
```

### 视频生成

```go
// 视频生成示例
func videoGeneration() {
    config, _ := aigcpanel.LoadConfigFromJSON("config.json")
    server := easyserver.NewEasyServer("/path/to/model", config)
    
    server.Start()
    defer server.Stop()
    
    result, err := server.VideoGen(easyserver.ServerFunctionDataType{
        ID:    "video-001",
        Audio: "/path/to/audio.wav",
        Video: "/path/to/reference.mp4",
        Param: map[string]interface{}{
            "fps":        30,
            "resolution": "1920x1080",
            "quality":    "high",
        },
    })
    
    if err != nil {
        log.Printf("视频生成失败: %v", err)
        return
    }
    
    fmt.Printf("视频生成结果: %+v\n", result)
}
```

### 语音识别 (ASR)

```go
// 语音识别示例
func speechRecognition() {
    config, _ := aigcpanel.LoadConfigFromJSON("config.json")
    server := easyserver.NewEasyServer("/path/to/model", config)
    
    server.Start()
    defer server.Stop()
    
    result, err := server.Asr(easyserver.ServerFunctionDataType{
        ID:    "asr-001",
        Audio: "/path/to/speech.wav",
        Param: map[string]interface{}{
            "language": "zh-CN",
            "model":    "whisper-large",
        },
    })
    
    if err != nil {
        log.Printf("语音识别失败: %v", err)
        return
    }
    
    if result.Code == 0 {
        fmt.Printf("识别结果: %v\n", result.Data)
    }
}
```

## 命令行工具

项目提供了一个命令行工具用于快速测试模型功能：

### 使用方法

```bash
# 编译命令行工具
go build -o model_caller ./examples/model_caller

# 使用语音合成功能
./model_caller /path/to/model /path/to/config.json soundTts text="你好，世界" speaker="中文女" speed=1.0

# 使用语音克隆功能
./model_caller /path/to/model /path/to/config.json soundClone text="克隆测试" promptAudio="/path/to/ref.wav" promptText="参考文本"

# 使用视频生成功能
./model_caller /path/to/model /path/to/config.json videoGen audio="/path/to/audio.wav" video="/path/to/ref.mp4"

# 使用语音识别功能
./model_caller /path/to/model /path/to/config.json asr audio="/path/to/speech.wav"
```

### 快速测试脚本

使用提供的测试脚本：

```bash
# 使用 run_model.sh 脚本进行测试
./run_model.sh /path/to/model /path/to/config.json soundTts text="你好，世界" speaker="中文女" speed=1.0
```

## 服务器状态管理

### 状态类型

```go
const (
    ServerStopped  ServerStatus = "stopped"  // 服务器已停止
    ServerStarting ServerStatus = "starting" // 服务器正在启动
    ServerRunning  ServerStatus = "running"  // 服务器正在运行
    ServerStopping ServerStatus = "stopping" // 服务器正在停止
    ServerError    ServerStatus = "error"    // 服务器出现错误
)
```

### 状态监控示例

```go
func monitorServerStatus(server *easyserver.EasyServer) {
    for {
        status := server.GetStatus()
        fmt.Printf("服务器状态: %s\n", status)
        
        switch status {
        case easyserver.ServerRunning:
            fmt.Println("服务器运行正常")
        case easyserver.ServerError:
            fmt.Println("服务器出现错误，需要重启")
            server.Stop()
            server.Start()
        }
        
        time.Sleep(5 * time.Second)
    }
}
```

## 错误处理

### 常见错误类型

1. **配置加载错误** - 检查配置文件格式和路径
2. **服务器启动失败** - 检查模型路径和依赖
3. **功能调用失败** - 检查参数格式和模型支持
4. **任务执行超时** - 调整超时设置或检查模型性能

### 错误处理示例

```go
func handleErrors() {
    config, err := aigcpanel.LoadConfigFromJSON("config.json")
    if err != nil {
        log.Printf("配置加载失败: %v", err)
        return
    }
    
    server := easyserver.NewEasyServer("/path/to/model", config)
    
    if err := server.Start(); err != nil {
        log.Printf("服务器启动失败: %v", err)
        return
    }
    defer func() {
        if err := server.Stop(); err != nil {
            log.Printf("服务器停止失败: %v", err)
        }
    }()
    
    result, err := server.SoundTts(easyserver.ServerFunctionDataType{
        ID:   "test",
        Text: "测试文本",
    })
    
    if err != nil {
        log.Printf("功能调用失败: %v", err)
        return
    }
    
    if result.Code != 0 {
        log.Printf("任务执行失败: %s", result.Msg)
        return
    }
    
    fmt.Printf("任务执行成功: %+v\n", result.Data)
}
```

## 最佳实践

### 1. 资源管理

```go
// 始终确保服务器正确关闭
defer server.Stop()

// 使用上下文控制超时
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
```

### 2. 并发安全

```go
// 使用互斥锁保护共享资源
var mu sync.Mutex
mu.Lock()
defer mu.Unlock()
```

### 3. 配置管理

```go
// 使用环境变量管理配置
modelPath := os.Getenv("MODEL_PATH")
if modelPath == "" {
    modelPath = "/default/model/path"
}
```

## 故障排除

### 常见问题

**Q: 服务器启动失败**
A: 检查模型路径是否正确，确保所有依赖已安装

**Q: 功能调用返回错误**
A: 验证参数格式，检查模型是否支持该功能

**Q: 任务执行超时**
A: 增加超时时间或检查模型性能和资源使用情况

**Q: 配置文件加载失败**
A: 验证 JSON 格式，检查文件路径和权限

### 调试技巧

1. 启用详细日志输出
2. 检查服务器状态变化
3. 验证输入参数格式
4. 监控系统资源使用

## 贡献

欢迎提交 Issue 和 Pull Request 来改进这个项目。

## 许可证

本项目采用 MIT 许可证。详见 LICENSE 文件。

## 更新日志

### v1.0.0
- 初始版本发布
- 支持语音合成、语音克隆、视频生成、语音识别功能
- 提供简易服务器管理和本地模型管理
- 包含完整的示例和文档
