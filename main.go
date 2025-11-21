package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// --------------------------- 配置常量 ---------------------------
const (
	DefaultPort           = 60001
	DiscoveryPort         = 60002
	DiscoveryResponsePort = 60003
	BufferSize            = 1024 * 1024 * 16
	TimeoutDuration       = 60 * time.Second
	DiscoveryMessage      = "GO_FILE_TRANSFER_DISCOVERY_REQUEST"
	DiscoveryResponse     = "GO_FILE_TRANSFER_DISCOVERY_RESPONSE"
	FileHeaderPrefix      = "FILE_START"
	EndMarker             = "TRANSFER_END"
	StatsMarker           = "STATS_INFO" // 新增：统计信息标记
)

// --------------------------- 传输统计结构体 ---------------------------
type TransferStats struct {
	TotalFiles       int     `json:"totalFiles"`       // 总文件数
	CompletedFiles   int     `json:"completedFiles"`   // 已完成文件数
	TotalBytes       int64   `json:"totalBytes"`       // 总字节数
	TransferredBytes int64   `json:"transferredBytes"` // 已传输字节数
	CurrentSpeed     float64 `json:"currentSpeed"`     // 当前传输速度 (MB/s)
	EstimatedTime    string  `json:"estimatedTime"`    // 预计剩余时间
	CurrentFile      string  `json:"currentFile"`      // 当前传输的文件名
	Progress         float64 `json:"progress"`         // 总体进度百分比 (0-100)
	Status           string  `json:"status"`           // 传输状态: "scanning", "transferring", "completed", "failed"
}

// --------------------------- 性能优化结构体 ---------------------------
type PerformanceStats struct {
	lastUpdateTime  time.Time     // 上次更新时间
	updateInterval  time.Duration // 更新间隔
	speedSamples    []float64     // 速度采样数组
	maxSpeedSamples int           // 最大采样数
	lastBytes       int64         // 上次字节数
}

// --------------------------- 应用结构体 ---------------------------
type App struct {
	ctx     context.Context
	mu      sync.Mutex
	Running bool             `json:"running"` // 是否正在收发
	Stats   TransferStats    `json:"stats"`   // 传输统计信息
	perf    PerformanceStats // 性能统计信息
}

// NewApp 创建新的App实例
func NewApp() *App {
	return &App{
		Running: false,
		perf: PerformanceStats{
			updateInterval:  200 * time.Millisecond, // 更新间隔200ms
			maxSpeedSamples: 10,                     // 最大速度采样数
			speedSamples:    make([]float64, 0, 10),
		},
	}
}

// --------------------------- 应用生命周期 ---------------------------
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.Running = false
}

// --------------------------- 工具方法 ---------------------------
func (a *App) emitStatusUpdate(status string) {
	wailsruntime.EventsEmit(a.ctx, "status-updated", status)
}

func (a *App) emitOperationCompleted() {
	wailsruntime.EventsEmit(a.ctx, "operation-completed")
}

func (a *App) emitStatsUpdated() {
	wailsruntime.EventsEmit(a.ctx, "stats-updated", a.Stats)
}

func (a *App) resetStats() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Stats = TransferStats{
		TotalFiles:       0,
		CompletedFiles:   0,
		TotalBytes:       0,
		TransferredBytes: 0,
		CurrentSpeed:     0,
		EstimatedTime:    "",
		CurrentFile:      "",
		Progress:         0,
		Status:           "ready",
	}
}

// --------------------------- 前端绑定方法 ---------------------------
func (a *App) GetStats() TransferStats {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.Stats
}

// GetFileInfo 获取文件/文件夹的详细信息
func (a *App) GetFileInfo(path string) map[string]interface{} {
	info := make(map[string]interface{})

	// 清理和验证路径
	cleanPath := strings.TrimSpace(path)
	if cleanPath == "" {
		info["error"] = "路径为空"
		return info
	}

	// 添加调试信息
	fmt.Printf("GetFileInfo 接收到的路径: '%s'\n", cleanPath)
	fmt.Printf("路径长度: %d\n", len(cleanPath))
	fmt.Printf("路径字符: %v\n", []rune(cleanPath))

	// 检查路径是否存在
	stat, err := os.Stat(cleanPath)
	if err != nil {
		// 提供更详细的错误信息
		if os.IsNotExist(err) {
			info["error"] = fmt.Sprintf("文件或文件夹不存在: %s", cleanPath)
			fmt.Printf("路径不存在错误: %v\n", err)
		} else if os.IsPermission(err) {
			info["error"] = fmt.Sprintf("没有权限访问: %s", cleanPath)
			fmt.Printf("权限错误: %v\n", err)
		} else {
			info["error"] = fmt.Sprintf("访问文件失败: %v", err)
			fmt.Printf("其他错误: %v\n", err)
		}
		return info
	}

	info["name"] = stat.Name()
	info["path"] = cleanPath
	info["isDirectory"] = stat.IsDir()
	info["size"] = stat.Size()
	info["modTime"] = stat.ModTime().Format("2006-01-02 15:04:05")

	if stat.IsDir() {
		// 如果是文件夹，计算总大小和文件数
		totalFiles, totalBytes, err := a.scanFiles(cleanPath)
		if err == nil {
			info["totalFiles"] = totalFiles
			info["totalBytes"] = totalBytes
			info["sizeDisplay"] = fmt.Sprintf("文件夹 (%d 个文件, %s)", totalFiles, formatFileSize(totalBytes))
		} else {
			info["sizeDisplay"] = "文件夹"
		}
	} else {
		// 如果是文件，直接显示大小
		info["sizeDisplay"] = formatFileSize(stat.Size())
	}

	fmt.Printf("GetFileInfo 成功处理路径: '%s', 类型: %v\n", cleanPath, stat.IsDir())
	return info
}

// formatFileSize 格式化文件大小显示
func formatFileSize(bytes int64) string {
	if bytes == 0 {
		return "0 B"
	}
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func (a *App) RestartReceive() error {
	a.mu.Lock()
	if a.Running {
		a.mu.Unlock()
		return fmt.Errorf("已有任务在进行")
	}
	a.Running = true
	a.mu.Unlock()

	go func() {
		defer func() {
			a.mu.Lock()
			a.Running = false
			a.mu.Unlock()
			a.emitOperationCompleted()
		}()

		a.emitStatusUpdate("正在接收...")
		a.receiver()
	}()

	return nil
}

func (a *App) SelectFile() string {
	filePath, err := wailsruntime.OpenFileDialog(a.ctx, wailsruntime.OpenDialogOptions{
		Title: "选择要发送的文件",
	})
	if err != nil {
		return ""
	}
	return filePath
}

func (a *App) SelectFolder() string {
	folderPath, err := wailsruntime.OpenDirectoryDialog(a.ctx, wailsruntime.OpenDialogOptions{
		Title: "选择要发送的文件夹",
	})
	if err != nil {
		return ""
	}
	return folderPath
}

func (a *App) Send(sourcePath string) error {
	a.mu.Lock()
	if a.Running {
		a.mu.Unlock()
		return fmt.Errorf("已有任务在进行")
	}
	a.Running = true
	a.mu.Unlock()

	// 使用通道等待传输完成
	done := make(chan bool, 1)

	go func() {
		defer func() {
			a.mu.Lock()
			a.Running = false
			a.mu.Unlock()
			a.emitOperationCompleted()
			done <- true
		}()

		a.emitStatusUpdate("正在扫描文件...")

		if _, err := os.Stat(sourcePath); err != nil {
			a.emitStatusUpdate("文件不存在: " + err.Error())
			return
		}

		targetIP, err := a.discoverTarget()
		if err != nil {
			a.emitStatusUpdate("发现接收端失败: " + err.Error())
			return
		}

		a.emitStatusUpdate("已连接到接收端: " + targetIP)
		a.emitStatusUpdate("正在传输文件...")
		a.sender(sourcePath, targetIP)
	}()

	// 等待传输开始（非阻塞）
	go func() {
		time.Sleep(100 * time.Millisecond)
		select {
		case <-done:
			// 传输已完成
		default:
			// 传输仍在进行
		}
	}()

	return nil
}

func (a *App) Receive() error {
	a.mu.Lock()
	if a.Running {
		a.mu.Unlock()
		return fmt.Errorf("已有任务在进行")
	}
	a.Running = true
	a.mu.Unlock()

	go func() {
		defer func() {
			a.mu.Lock()
			a.Running = false
			a.mu.Unlock()
			a.emitOperationCompleted()
		}()

		a.emitStatusUpdate("正在接收...")
		a.receiver()
	}()

	return nil
}

// --------------------------- 网络工具方法 ---------------------------
func getLocalIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		ifaces, _ := net.Interfaces()
		for _, iface := range ifaces {
			if iface.Flags&net.FlagUp != 0 && iface.Flags&net.FlagLoopback == 0 {
				addrs, _ := iface.Addrs()
				for _, addr := range addrs {
					if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
						return ipNet.IP.String(), nil
					}
				}
			}
		}
		return "", fmt.Errorf("无法获取本地IP")
	}
	defer conn.Close()
	return conn.LocalAddr().(*net.UDPAddr).IP.String(), nil
}

// --------------------------- 自动发现 ---------------------------
func (a *App) discoverTarget() (string, error) {
	localIP, err := getLocalIP()
	if err != nil {
		return "", fmt.Errorf("获取本地IP失败: %v", err)
	}

	localAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", localIP, DiscoveryResponsePort))
	if err != nil {
		return "", err
	}
	conn, err := net.ListenUDP("udp", localAddr)
	if err != nil {
		return "", fmt.Errorf("监听 UDP 端口失败: %v", err)
	}
	defer conn.Close()
	conn.SetReadDeadline(time.Now().Add(TimeoutDuration))

	broadcastAddr, _ := net.ResolveUDPAddr("udp", fmt.Sprintf("255.255.255.255:%d", DiscoveryPort))
	req := []byte(fmt.Sprintf("%s|%s|%d", DiscoveryMessage, localIP, DiscoveryResponsePort))
	for i := 0; i < 3; i++ {
		conn.WriteToUDP(req, broadcastAddr)
		time.Sleep(time.Second)
	}

	buf := make([]byte, 1024)
	n, remoteAddr, err := conn.ReadFromUDP(buf)
	if err != nil {
		return "", fmt.Errorf("未发现接收端或超时 (%v)", err)
	}
	if strings.TrimSpace(string(buf[:n])) == DiscoveryResponse {
		return remoteAddr.IP.String(), nil
	}
	return "", fmt.Errorf("收到无效响应")
}

func (a *App) handleDiscovery(quit chan struct{}) {
	localIP, err := getLocalIP()
	if err != nil {
		return
	}
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", localIP, DiscoveryPort))
	if err != nil {
		return
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return
	}
	defer conn.Close()

	buf := make([]byte, 1024)
	for {
		select {
		case <-quit:
			return
		default:
			conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
			n, _, err := conn.ReadFromUDP(buf)
			if err != nil {
				continue
			}
			parts := strings.Split(strings.TrimSpace(string(buf[:n])), "|")
			if len(parts) == 3 && parts[0] == DiscoveryMessage {
				senderIP := parts[1]
				senderPort, _ := strconv.Atoi(parts[2])
				respAddr, _ := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", senderIP, senderPort))
				conn.WriteToUDP([]byte(DiscoveryResponse), respAddr)
			}
		}
	}
}

// --------------------------- 文件扫描和统计工具 ---------------------------
func (a *App) scanFiles(path string) (int, int64, error) {
	var totalFiles int
	var totalBytes int64

	stat, err := os.Stat(path)
	if err != nil {
		return 0, 0, err
	}

	if !stat.IsDir() {
		return 1, stat.Size(), nil
	}

	err = filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			totalFiles++
			totalBytes += info.Size()
		}
		return nil
	})

	return totalFiles, totalBytes, err
}

// --------------------------- 优化的统计更新方法 ---------------------------
func (a *App) updateStatsOptimized(currentFile string, fileSize int64, transferredBytes int64, startTime time.Time) {
	now := time.Now()

	// 检查是否需要更新（避免频繁更新导致的性能问题）
	if now.Sub(a.perf.lastUpdateTime) < a.perf.updateInterval {
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	// 更新当前文件（只显示总体进度，不显示单个文件进度）
	if currentFile != "" {
		a.Stats.CurrentFile = "传输中..."
	}

	// 更新传输字节数
	a.Stats.TransferredBytes = transferredBytes

	// 计算进度（基于总字节数）
	if a.Stats.TotalBytes > 0 {
		// 确保进度计算正确，避免除零错误
		progress := float64(transferredBytes) / float64(a.Stats.TotalBytes) * 100

		// 限制进度范围在0-100之间
		if progress < 0 {
			a.Stats.Progress = 0
		} else if progress > 100 {
			a.Stats.Progress = 100
		} else {
			a.Stats.Progress = progress
		}
	} else {
		// 如果总字节数为0，设置一个小的初始进度
		a.Stats.Progress = 0.1
	}

	// 计算传输速度（使用滑动窗口平均）
	elapsed := now.Sub(startTime).Seconds()
	if elapsed > 0.1 { // 至少需要0.1秒才能计算有效速度
		currentSpeed := float64(transferredBytes) / (1024 * 1024) / elapsed // MB/s

		// 添加速度采样
		a.perf.speedSamples = append(a.perf.speedSamples, currentSpeed)
		if len(a.perf.speedSamples) > a.perf.maxSpeedSamples {
			a.perf.speedSamples = a.perf.speedSamples[1:]
		}

		// 计算平均速度（加权平均，最近的速度权重更高）
		var totalSpeed float64
		var totalWeight float64
		for i, speed := range a.perf.speedSamples {
			weight := float64(i + 1) // 越新的速度权重越高
			totalSpeed += speed * weight
			totalWeight += weight
		}
		if totalWeight > 0 {
			a.Stats.CurrentSpeed = totalSpeed / totalWeight
		} else {
			a.Stats.CurrentSpeed = currentSpeed
		}
	} else {
		// 传输刚开始，使用瞬时速度
		if transferredBytes > 0 {
			a.Stats.CurrentSpeed = float64(transferredBytes) / (1024 * 1024) / elapsed
		} else {
			a.Stats.CurrentSpeed = 0
		}
	}

	// 计算预计剩余时间（使用加权平均速度）
	if a.Stats.CurrentSpeed > 0 && a.Stats.TotalBytes > 0 {
		remainingBytes := a.Stats.TotalBytes - transferredBytes
		remainingSeconds := float64(remainingBytes) / (a.Stats.CurrentSpeed * 1024 * 1024)

		// 平滑剩余时间计算，避免剧烈波动
		if remainingSeconds < 1 {
			a.Stats.EstimatedTime = "<1秒"
		} else if remainingSeconds < 60 {
			a.Stats.EstimatedTime = fmt.Sprintf("%.0f秒", remainingSeconds)
		} else if remainingSeconds < 3600 {
			a.Stats.EstimatedTime = fmt.Sprintf("%.1f分钟", remainingSeconds/60)
		} else {
			a.Stats.EstimatedTime = fmt.Sprintf("%.1f小时", remainingSeconds/3600)
		}
	} else {
		a.Stats.EstimatedTime = "计算中..."
	}

	// 更新性能统计
	a.perf.lastUpdateTime = now
	a.perf.lastBytes = transferredBytes

	// 触发前端更新
	a.emitStatsUpdated()
}

// --------------------------- 文件传输进度更新 ---------------------------
func (a *App) updateStats(currentFile string, fileSize int64, transferredBytes int64, startTime time.Time) {
	// 使用优化版本
	a.updateStatsOptimized(currentFile, fileSize, transferredBytes, startTime)
}

// --------------------------- 发送 / 接收 逻辑 ---------------------------
func (a *App) sendFileOrFolder(conn net.Conn, rootPath, baseDir string, startTime time.Time, transferredBytes *int64) error {
	fi, err := os.Stat(rootPath)
	if err != nil {
		return fmt.Errorf("获取文件信息失败 %s: %v", rootPath, err)
	}

	if !fi.IsDir() {
		rel, _ := filepath.Rel(baseDir, rootPath)
		if rel == "." {
			rel = filepath.Base(rootPath)
		}

		// 更新当前文件状态
		a.updateStats(rel, fi.Size(), *transferredBytes, startTime)

		// 使用defer确保文件句柄正确关闭
		f, err := os.Open(rootPath)
		if err != nil {
			return fmt.Errorf("打开文件失败 %s: %v", rootPath, err)
		}
		defer func() {
			if closeErr := f.Close(); closeErr != nil {
				fmt.Printf("关闭文件失败 %s: %v\n", rootPath, closeErr)
			}
		}()

		// 发送文件头
		hdr := fmt.Sprintf("%s|%s|%d\n", FileHeaderPrefix, filepath.ToSlash(rel), fi.Size())
		// 设置写入超时，避免网络阻塞
		conn.SetWriteDeadline(time.Now().Add(30 * time.Second))
		if _, err = conn.Write([]byte(hdr)); err != nil {
			return fmt.Errorf("发送文件头失败 %s: %v", rel, err)
		}
		// 重置写入超时
		conn.SetWriteDeadline(time.Time{})

		// 发送文件内容并实时更新进度
		buffer := make([]byte, BufferSize)
		var totalWritten int64
		for {
			n, err := f.Read(buffer)
			if n > 0 {
				// 设置写入超时，避免网络阻塞
				conn.SetWriteDeadline(time.Now().Add(30 * time.Second))
				written, err := conn.Write(buffer[:n])
				if err != nil {
					return fmt.Errorf("发送文件内容失败 %s: %v", rel, err)
				}
				// 重置写入超时
				conn.SetWriteDeadline(time.Time{})

				totalWritten += int64(written)
				*transferredBytes += int64(written)

				// 实时更新统计信息（优化更新频率）
				if totalWritten%int64(BufferSize*10) == 0 || totalWritten == fi.Size() {
					a.updateStats(rel, fi.Size(), *transferredBytes, startTime)
				}
			}
			if err == io.EOF {
				break
			}
			if err != nil {
				return fmt.Errorf("读取文件失败 %s: %v", rel, err)
			}
		}

		// 确保文件传输完成时更新统计
		a.Stats.CompletedFiles++
		a.updateStats("", fi.Size(), *transferredBytes, startTime)

		return nil
	}

	// 处理文件夹
	entries, err := os.ReadDir(rootPath)
	if err != nil {
		return fmt.Errorf("读取目录失败 %s: %v", rootPath, err)
	}

	if rootPath == baseDir {
		baseDir = filepath.Dir(rootPath)
	}

	for _, e := range entries {
		fullPath := filepath.Join(rootPath, e.Name())
		if err = a.sendFileOrFolder(conn, fullPath, baseDir, startTime, transferredBytes); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) sender(sourcePath, targetIP string) {
	if _, err := os.Stat(sourcePath); err != nil {
		return
	}

	// 重置统计信息
	a.resetStats()

	// 扫描文件获取总数和总大小
	a.mu.Lock()
	a.Stats.Status = "scanning"
	a.emitStatsUpdated()
	a.mu.Unlock()

	totalFiles, totalBytes, err := a.scanFiles(sourcePath)
	if err != nil {
		return
	}

	// 更新统计信息
	a.mu.Lock()
	a.Stats.TotalFiles = totalFiles
	a.Stats.TotalBytes = totalBytes
	a.Stats.Status = "transferring"
	a.emitStatsUpdated()
	a.mu.Unlock()

	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", targetIP, DefaultPort), TimeoutDuration)
	if err != nil {
		return
	}
	defer conn.Close()

	fi, _ := os.Stat(sourcePath)
	rootName := fi.Name()
	isDirFlag := "FILE"
	if fi.IsDir() {
		isDirFlag = "DIR"
	}
	metaData := fmt.Sprintf("%s|%s\n", rootName, isDirFlag)
	if _, err = conn.Write([]byte(metaData)); err != nil {
		return
	}

	// 发送统计信息给接收方，确保接收方有正确的进度计算基础
	statsData := fmt.Sprintf("%s|%d|%d\n", StatsMarker, totalFiles, totalBytes)
	if _, err = conn.Write([]byte(statsData)); err != nil {
		return
	}

	baseDir := filepath.Dir(sourcePath)
	if !fi.IsDir() {
		baseDir = sourcePath
	}

	startTime := time.Now()
	var transferredBytes int64

	if err = a.sendFileOrFolder(conn, sourcePath, baseDir, startTime, &transferredBytes); err != nil {
	} else {
		conn.Write([]byte(EndMarker + "\n"))

		// 传输完成
		a.mu.Lock()
		a.Stats.Status = "completed"
		a.Stats.Progress = 100
		a.Stats.CompletedFiles = a.Stats.TotalFiles
		a.Stats.TransferredBytes = a.Stats.TotalBytes
		a.emitStatsUpdated()
		a.mu.Unlock()
	}
}

func (a *App) receiver() {
	localIP, err := getLocalIP()
	if err != nil {
		a.emitStatusUpdate("获取本地IP失败: " + err.Error())
		return
	}

	// 重置统计信息
	a.resetStats()
	a.mu.Lock()
	a.Stats.Status = "waiting"
	a.emitStatsUpdated()
	a.mu.Unlock()

	a.emitStatusUpdate("正在等待连接...")

	quit := make(chan struct{})
	go a.handleDiscovery(quit)

	ln, err := net.Listen("tcp", fmt.Sprintf("%s:%d", localIP, DefaultPort))
	if err != nil {
		close(quit)
		a.emitStatusUpdate("监听端口失败: " + err.Error())
		return
	}
	defer ln.Close()

	// 设置超时，避免无限等待
	ln.(*net.TCPListener).SetDeadline(time.Now().Add(TimeoutDuration * 5))

	a.emitStatusUpdate("等待发送方连接...")

	conn, err := ln.Accept()
	close(quit)
	if err != nil {
		a.emitStatusUpdate("接受连接失败: " + err.Error())
		return
	}
	defer conn.Close()

	a.emitStatusUpdate("已连接到发送方，开始接收...")

	conn.SetReadDeadline(time.Now().Add(TimeoutDuration))
	reader := bufio.NewReader(conn)
	metaData, err := reader.ReadString('\n')
	if err != nil {
		a.emitStatusUpdate("读取元数据失败: " + err.Error())
		return
	}
	conn.SetReadDeadline(time.Time{})
	parts := strings.Split(strings.TrimSpace(metaData), "|")
	if len(parts) != 2 {
		a.emitStatusUpdate("元数据格式错误")
		return
	}
	rootName, isDirFlag := parts[0], parts[1]
	if isDirFlag == "DIR" {
		os.MkdirAll(rootName, 0755)
	}

	// 接收统计信息
	statsData, err := reader.ReadString('\n')
	if err != nil {
		a.emitStatusUpdate("读取统计信息失败: " + err.Error())
		return
	}
	statsParts := strings.Split(strings.TrimSpace(statsData), "|")
	if len(statsParts) == 3 && statsParts[0] == StatsMarker {
		// 使用发送方提供的统计信息初始化接收方统计
		totalFiles, _ := strconv.Atoi(statsParts[1])
		totalBytes, _ := strconv.ParseInt(statsParts[2], 10, 64)

		a.mu.Lock()
		a.Stats.TotalFiles = totalFiles
		a.Stats.TotalBytes = totalBytes
		a.Stats.Progress = 0.1 // 设置初始进度为0.1%，避免显示0%
		a.Stats.Status = "transferring"
		a.emitStatsUpdated()
		a.mu.Unlock()
	} else {
		// 向后兼容：如果没有收到统计信息，使用默认值
		a.mu.Lock()
		a.Stats.TotalFiles = 1
		a.Stats.TotalBytes = 1
		a.Stats.Progress = 0.1 // 设置初始进度为0.1%，避免显示0%
		a.Stats.Status = "transferring"
		a.emitStatsUpdated()
		a.mu.Unlock()
	}

	startTime := time.Now()
	var receivedBytes int64
	var completedFiles int

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			a.emitStatusUpdate("读取文件头失败: " + err.Error())
			break
		}
		line = strings.TrimSpace(line)
		if line == EndMarker {
			a.emitStatusUpdate("传输完成")
			break
		}
		if !strings.HasPrefix(line, FileHeaderPrefix) {
			a.emitStatusUpdate("无效的文件头格式")
			break
		}
		hdr := strings.Split(line, "|")
		if len(hdr) != 3 {
			a.emitStatusUpdate("文件头格式错误")
			break
		}
		relPath := hdr[1]
		fileSize, _ := strconv.ParseInt(hdr[2], 10, 64)
		targetPath := filepath.Join(rootName, filepath.FromSlash(relPath))
		os.MkdirAll(filepath.Dir(targetPath), 0755)

		// 更新当前文件状态
		a.updateStats(relPath, fileSize, receivedBytes, startTime)

		// 创建文件
		file, err := os.Create(targetPath)
		if err != nil {
			a.emitStatusUpdate("创建文件失败: " + err.Error())
			break
		}

		// 接收文件内容并实时更新进度
		buffer := make([]byte, BufferSize)
		var totalReceived int64
		var fileWriteError error

		for totalReceived < fileSize && fileWriteError == nil {
			remaining := fileSize - totalReceived
			if remaining > int64(len(buffer)) {
				remaining = int64(len(buffer))
			}

			// 设置读取超时，避免网络阻塞
			conn.SetReadDeadline(time.Now().Add(30 * time.Second))
			n, err := reader.Read(buffer[:remaining])
			if n > 0 {
				// 重置读取超时
				conn.SetReadDeadline(time.Time{})

				written, err := file.Write(buffer[:n])
				if err != nil {
					fileWriteError = err
					break
				}
				totalReceived += int64(written)
				receivedBytes += int64(written)

				// 实时更新统计信息（优化更新频率）
				if totalReceived%int64(BufferSize*10) == 0 || totalReceived == fileSize {
					a.updateStats(relPath, fileSize, receivedBytes, startTime)
				}
			}
			if err != nil {
				if err == io.EOF {
					break
				}
				a.emitStatusUpdate("读取文件内容失败: " + err.Error())
				break
			}
		}

		// 确保文件正确关闭
		if closeErr := file.Close(); closeErr != nil {
			fmt.Printf("关闭文件失败 %s: %v\n", targetPath, closeErr)
		}

		// 如果文件写入失败，删除不完整的文件
		if fileWriteError != nil {
			os.Remove(targetPath)
			a.emitStatusUpdate("写入文件失败: " + fileWriteError.Error())
			break
		}

		completedFiles++

		// 更新统计信息 - 接收端动态调整总数
		a.mu.Lock()
		a.Stats.CompletedFiles = completedFiles
		a.Stats.TransferredBytes = receivedBytes
		// 动态调整总文件数，使用已完成的文件数作为参考
		if completedFiles > a.Stats.TotalFiles {
			a.Stats.TotalFiles = completedFiles
		}
		// 动态调整总字节数，使用已接收的字节数作为参考
		if receivedBytes > a.Stats.TotalBytes {
			a.Stats.TotalBytes = receivedBytes
		}
		a.emitStatsUpdated()
		a.mu.Unlock()
	}

	// 传输完成
	a.mu.Lock()
	a.Stats.Status = "completed"
	a.Stats.Progress = 100
	a.Stats.CompletedFiles = completedFiles
	a.Stats.TransferredBytes = receivedBytes
	a.Stats.TotalFiles = completedFiles
	a.Stats.TotalBytes = receivedBytes
	a.emitStatsUpdated()
	a.mu.Unlock()

	a.emitStatusUpdate("文件接收完成")
}
