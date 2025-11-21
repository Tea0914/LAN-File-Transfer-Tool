# 局域网文件传输工具 / LAN File Transfer Tool

[English](#english) | [中文](#chinese)

---

<a name="english"></a>
# LAN File Transfer Tool

A high-performance, cross-platform file transfer application built with Go and Wails framework, designed for fast and reliable file sharing within local area networks.

## Features

- 🚀 **High-speed Transfer**: Optimized for large file transfers with 16MB buffer size
- 📊 **Real-time Progress**: Live transfer statistics including speed, progress, and estimated time
- 🔍 **Auto Discovery**: Automatic device discovery within the same network
- 📁 **File & Folder Support**: Transfer both individual files and entire folders
- 🎯 **Cross-platform**: Built with Wails for Windows, macOS, and Linux compatibility
- 📈 **Performance Monitoring**: Real-time speed calculation and progress tracking
- 🔄 **Reliable Transfer**: Robust error handling and connection management

## Technology Stack

- **Backend**: Go 1.24.2
- **Frontend**: Vite + Vanilla JavaScript
- **Framework**: Wails v2.11.0
- **Networking**: TCP/UDP for file transfer and device discovery

## Installation

### Prerequisites

- Go 1.24.2 or later
- Node.js 18+ and npm
- Wails CLI: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`

### Build from Source

1. Clone the repository:
```bash
git clone <repository-url>
cd LAN-File-Transfer-Tool
```

2. Install dependencies:
```bash
cd frontend
npm install
cd ..
```

3. Build the application:
```bash
wails build
```

4. Run the development version:
```bash
wails dev
```

## Usage

### Sending Files

1. **Start the application** on both sending and receiving devices
2. **On the sending device**:
   - Click "Select File" or "Select Folder" to choose files
   - Click "Send" to initiate transfer
   - The app will automatically discover the receiving device

3. **On the receiving device**:
   - Click "Receive" to start listening for incoming transfers
   - The app will automatically accept the connection

### Network Requirements

- Both devices must be on the same local network
- Firewall should allow connections on ports 60001-60003
- No internet connection required

### Port Configuration

- **File Transfer**: Port 60001 (TCP)
- **Device Discovery**: Port 60002 (UDP)
- **Discovery Response**: Port 60003 (UDP)

## Development

### Project Structure

```
LAN-File-Transfer-Tool/
├── main.go              # Main application entry point
├── app.go               # Core application logic
├── wails.json           # Wails configuration
├── go.mod               # Go module dependencies
├── frontend/            # Frontend application
│   ├── src/
│   │   └── main.js      # Frontend entry point
│   ├── index.html       # Main HTML file
│   └── package.json     # Frontend dependencies
└── build/               # Build artifacts and resources
```

### Key Components

- **File Transfer**: TCP-based reliable file transfer
- **Device Discovery**: UDP-based automatic device detection
- **Progress Tracking**: Real-time statistics and progress updates
- **Error Handling**: Comprehensive error management and recovery

## Performance Features

- **Large Buffer Size**: 16MB buffer for efficient large file transfers
- **Optimized Updates**: Smart progress update intervals to reduce overhead
- **Speed Calculation**: Weighted average speed calculation for accuracy
- **Memory Efficient**: Stream-based processing for low memory usage

## Troubleshooting

### Common Issues

1. **Devices not discovering each other**
   - Ensure both devices are on the same network
   - Check firewall settings for ports 60001-60003
   - Verify network connectivity

2. **Transfer fails or is slow**
   - Check available disk space on receiving device
   - Ensure stable network connection
   - Try transferring smaller files first

3. **Application won't start**
   - Verify all dependencies are installed
   - Check Wails installation with `wails doctor`

## License

This project is licensed under the terms included in the LICENSE file.

---

<a name="chinese"></a>
# 局域网文件传输工具

一个使用Go和Wails框架构建的高性能跨平台文件传输应用程序，专为局域网内快速可靠的文件共享而设计。

## 功能特性

- 🚀 **高速传输**: 针对大文件传输优化，使用16MB缓冲区
- 📊 **实时进度**: 实时传输统计，包括速度、进度和预计时间
- 🔍 **自动发现**: 同一网络内自动发现设备
- 📁 **文件与文件夹支持**: 支持传输单个文件和整个文件夹
- 🎯 **跨平台**: 使用Wails构建，支持Windows、macOS和Linux
- 📈 **性能监控**: 实时速度计算和进度跟踪
- 🔄 **可靠传输**: 强大的错误处理和连接管理

## 技术栈

- **后端**: Go 1.24.2
- **前端**: Vite + 原生JavaScript
- **框架**: Wails v2.11.0
- **网络**: TCP/UDP用于文件传输和设备发现

## 安装

### 环境要求

- Go 1.24.2 或更高版本
- Node.js 18+ 和 npm
- Wails CLI: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`

### 从源码构建

1. 克隆仓库:
```bash
git clone <仓库地址>
cd LAN-File-Transfer-Tool
```

2. 安装依赖:
```bash
cd frontend
npm install
cd ..
```

3. 构建应用程序:
```bash
wails build
```

4. 运行开发版本:
```bash
wails dev
```

## 使用方法

### 发送文件

1. **在两台设备上启动应用程序**
2. **在发送设备上**:
   - 点击"选择文件"或"选择文件夹"选择文件
   - 点击"发送"开始传输
   - 应用程序会自动发现接收设备

3. **在接收设备上**:
   - 点击"接收"开始监听传入的传输
   - 应用程序会自动接受连接

### 网络要求

- 两台设备必须在同一局域网内
- 防火墙应允许端口60001-60003的连接
- 不需要互联网连接

### 端口配置

- **文件传输**: 端口 60001 (TCP)
- **设备发现**: 端口 60002 (UDP)
- **发现响应**: 端口 60003 (UDP)

## 开发

### 项目结构

```
LAN-File-Transfer-Tool/
├── main.go              # 主应用程序入口
├── app.go               # 核心应用逻辑
├── wails.json           # Wails配置
├── go.mod               # Go模块依赖
├── frontend/            # 前端应用
│   ├── src/
│   │   └── main.js      # 前端入口
│   ├── index.html       # 主HTML文件
│   └── package.json     # 前端依赖
└── build/               # 构建产物和资源
```

### 核心组件

- **文件传输**: 基于TCP的可靠文件传输
- **设备发现**: 基于UDP的自动设备检测
- **进度跟踪**: 实时统计和进度更新
- **错误处理**: 全面的错误管理和恢复

## 性能特性

- **大缓冲区**: 16MB缓冲区用于高效大文件传输
- **优化更新**: 智能进度更新间隔以减少开销
- **速度计算**: 加权平均速度计算确保准确性
- **内存高效**: 基于流的处理，内存使用低

## 故障排除

### 常见问题

1. **设备无法相互发现**
   - 确保两台设备在同一网络
   - 检查防火墙设置，确保端口60001-60003开放
   - 验证网络连接性

2. **传输失败或速度慢**
   - 检查接收设备的可用磁盘空间
   - 确保网络连接稳定
   - 先尝试传输较小的文件

3. **应用程序无法启动**
   - 验证所有依赖是否已安装
   - 使用 `wails doctor` 检查Wails安装

## 许可证

本项目根据LICENSE文件中包含的条款进行许可。
