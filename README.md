# AIGCPanel

![](./screenshots/cn/home.png)

![star](https://img.shields.io/badge/Framework-TS+Vue3+Electron-blue)
[![star](https://img.shields.io/badge/WEB-aigcpanel.com-blue)](https://aigcpanel.com)
[![star](https://img.shields.io/github/stars/modstart-lib/aigcpanel.svg)](https://github.com/modstart-lib/aigcpanel)
[![star](https://gitee.com/modstart-lib/aigcpanel/badge/star.svg)](https://gitee.com/modstart-lib/aigcpanel)
[![star](https://gitcode.com/modstart-lib/aigcpanel/star/badge.svg)](https://gitcode.com/modstart-lib/aigcpanel)

## 软件介绍

`AIGCPanel` 是一款简单易用的一站式AI数字人系统，即使是小白用户也能轻松上手。它集成了视频合成、语音合成、语音克隆等核心功能，大大简化了本地AI模型的管理流程，支持一键导入和使用各种AI模型，让AI创作变得触手可及。

> **重要提醒**：禁止使用本产品进行任何违法违规业务，请在使用本软件时严格遵守中华人民共和国法律法规。

## 目录

- [功能特性](#功能特性)
- [模型支持](#模型支持)
- [功能预览](#功能预览)
- [安装使用](#安装使用)
- [技术栈](#技术栈)
- [本地运行开发](#本地运行开发)
- [加入交流群](#加入交流群)
- [License](#license)

## 功能特性

- **视频数字人合成**：支持视频画面与声音的精准换口型匹配，打造逼真数字人效果
- **语音处理**：提供语音合成、语音克隆、视频声音替换功能，支持多种声音参数自定义设置
- **模型管理**：支持多模型导入、一键启动、模型配置、日志查看等便捷操作
- **国际化支持**：内置简体中文和英语界面，满足不同用户需求
- **一键启动包**：提供多种模型的一键启动包，快速部署和使用

## 模型支持

### 声音合成

- 声音模型 [CosyVoice-300M](https://github.com/FunAudioLLM/CosyVoice)
- 声音模型 [CosyVoice-300M-Instruct](https://github.com/FunAudioLLM/CosyVoice)
- 声音模型 [CosyVoice2-0.5b](https://github.com/FunAudioLLM/CosyVoice)
- 声音模型 [FishSpeech](https://github.com/fishaudio/fish-speech)
- 声音模型 [IndexTTS](https://github.com/index-tts/index-tts)
- 声音模型 [SparkTTS](https://github.com/SparkAudio/Spark-TTS)
- 声音模型 [GPT-SoVITS](https://github.com/RVC-Boss/GPT-SoVITS)

### 声音识别

- 声音识别模型 [FunASR](https://github.com/modelscope/FunASR)

### 视频模型

- 视频模型 [MuseTalk](https://github.com/TMElyralab/MuseTalk)
- 视频模型 [LatentSync](https://github.com/bytedance/LatentSync)
- 视频模型 [Wav2Lip](https://github.com/Rudrabha/Wav2Lip)
- 视频模型 [Heygem](https://github.com/duixcom/Duix.Heygem)

## 功能预览

### 视频合成

![](./screenshots/cn/video-gen.png)

### 我的形象

![](./screenshots/cn/video-template.png)

### 语音克隆

![](./screenshots/cn/sound-clone.png)

### 语音合成

![](./screenshots/cn/sound-tts.png)

### 模型管理

![](./screenshots/cn/server.png)

## 安装使用

### Windows

- 访问 [https://aigcpanel.com](https://aigcpanel.com) 下载 Windows 安装包，一键安装即可

安装完成后，打开软件，下载模型一键启动包，即可使用。

### macOS

- 访问 [https://aigcpanel.com](https://aigcpanel.com) 下载 macOS 安装包（支持 Intel 和 Apple Silicon）
- 双击安装包进行安装
- 首次运行可能需要允许来自未知开发者的应用

安装完成后，打开软件，下载模型一键启动包，即可使用。

### Linux

- 访问 [https://aigcpanel.com](https://aigcpanel.com) 下载 Linux 安装包（AppImage 或 deb 格式）
- 对于 AppImage：给予执行权限后直接运行
- 对于 deb：使用包管理器安装

安装完成后，打开软件，下载模型一键启动包，即可使用。

## 技术栈

本项目基于以下技术栈构建：

- **Electron**: 跨平台桌面应用框架
- **Vue 3**: 现代前端框架
- **TypeScript**: 类型安全的JavaScript超集

## 本地运行开发

> 仅在 node 20 测试过

```shell
# 安装依赖
npm install
# 调试运行
npm run dev
# 打包
npm run build
```


## Go 重写版本（接口 + 入库）

为支持“使用 Golang 重写项目接口与入库逻辑”，仓库已提供 Go 后端骨架：

```bash
# 启动 Go 服务（默认 :8080）
# 默认 JSON 文件库: go/data/aigcpanel.json（在 `go` 目录下运行时为 `data/aigcpanel.json`）
cd go
go run ./cmd/aigcpaneld
```

可配置环境变量：

- `AIGCPANEL_ADDR`：监听地址（默认 `:8080`）
- `AIGCPANEL_DSN`：数据库文件路径（默认 `data/aigcpanel.json`）

当前已提供的接口（均为持久化入库逻辑）：

- `GET /health`
- `GET/POST /api/v1/users`
- `GET/POST /api/v1/servers`
- `PATCH /api/v1/servers/{id}`
- `GET/POST /api/v1/tasks`
- `PATCH /api/v1/tasks/{id}`

该实现用于作为 Go 全量重写的主干：接口层、仓储层、文件入库（JSON）已打通，可在此基础上继续对接模型推理、任务编排、鉴权与文件管理。

## 贡献指南

我们欢迎社区贡献！如果您想为 AIGCPanel 做出贡献，请：

1. Fork 本项目
2. 创建您的特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交您的更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建一个 Pull Request

## 问题报告

如果您遇到问题或有建议，请：

- 查看 [Issues](../../issues) 页面，确认问题是否已被报告
- 如果没有，请创建一个新的 Issue，并提供详细的描述和系统信息
- 对于技术支持，请加入我们的交流群

## 更新日志

查看 [CHANGELOG.md](./changelog.md) 了解最新更新和修复内容。

## 加入交流群

> 添加好友请备注 AIGCPanel

<table width="100%">
    <thead>
        <tr>
            <th width="50%">微信交流群</th>
            <th>QQ交流群</th>
        </tr>
    </thead>
    <tbody>
        <tr>
            <td>
                <img style="width:100%;"
                     src="https://modstart.com/code_dynamic/modstart_wx" />
            </td>
            <td>
                <img style="width:100%;" 
                     src="https://modstart.com/code_dynamic/modstart_qq" />
            </td>
        </tr>
    </tbody>
</table>

## License

Apache-2.0
