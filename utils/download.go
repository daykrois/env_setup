package utils

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/cheggaaa/pb/v3"
)

// Download 下载文件并显示进度条
func Download(url, path string) error {
	// 创建 HTTP 请求
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("下载出错: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("服务器返回错误状态码: %d", resp.StatusCode)
	}

	// 创建目标文件
	out, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("创建文件出错: %w", err)
	}
	defer out.Close()

	// 初始化进度条
	totalBytes := resp.ContentLength
	bar := pb.Full.Start64(totalBytes)
	defer bar.Finish()

	// 使用缓冲 I/O 提高性能
	bufferedReader := bufio.NewReader(resp.Body)
	bufferedWriter := bufio.NewWriter(out)
	reader := bar.NewProxyReader(bufferedReader)

	// 写入文件
	_, err = io.Copy(bufferedWriter, reader)
	if err != nil {
		return fmt.Errorf("写入文件出错: %w", err)
	}

	// 刷新缓冲区
	if err := bufferedWriter.Flush(); err != nil {
		return fmt.Errorf("刷新缓冲区出错: %w", err)
	}

	return nil
}
