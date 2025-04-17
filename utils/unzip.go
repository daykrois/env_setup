package utils

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/cheggaaa/pb/v3"
)

func UnzipWithProgress(src, dest string) error {
	// 打开 ZIP 文件
	r, err := zip.OpenReader(src)
	if err != nil {
		return fmt.Errorf("无法打开 ZIP 文件: %w", err)
	}
	defer r.Close()

	// 计算总大小
	var totalSize int64
	for _, f := range r.File {
		totalSize += int64(f.UncompressedSize64)
	}

	// 创建进度条
	bar := pb.Full.Start64(totalSize)
	defer bar.Finish()

	// 并发解压
	var wg sync.WaitGroup
	errChan := make(chan error, len(r.File))
	currentSize := int64(0)

	for _, f := range r.File {
		wg.Add(1)
		go func(f *zip.File) {
			defer wg.Done()

			// 构造目标路径
			fpath := filepath.Join(dest, f.Name)

			// 如果是目录，直接创建
			if f.FileInfo().IsDir() {
				if err := os.MkdirAll(fpath, os.ModePerm); err != nil {
					errChan <- fmt.Errorf("创建目录失败: %w", err)
					return
				}
				return
			}

			// 确保父目录存在
			if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
				errChan <- fmt.Errorf("创建父目录失败: %w", err)
				return
			}

			// 打开 ZIP 文件中的文件
			rc, err := f.Open()
			if err != nil {
				errChan <- fmt.Errorf("打开 ZIP 中的文件失败: %w", err)
				return
			}
			defer rc.Close()

			// 创建目标文件
			outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				errChan <- fmt.Errorf("创建目标文件失败: %w", err)
				return
			}
			defer outFile.Close()

			// 使用带进度的 Writer
			writer := bar.NewProxyWriter(outFile)
			n, err := io.Copy(writer, rc)
			if err != nil {
				errChan <- fmt.Errorf("写入文件失败: %w", err)
				return
			}

			// 更新当前解压的总字节数
			atomicAdd(&currentSize, n)
		}(f)
	}

	// 等待所有 Goroutines 完成
	wg.Wait()
	close(errChan)

	// 检查是否有错误
	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}

// 原子操作：增加 currentSize
func atomicAdd(currentSize *int64, delta int64) {
	*currentSize += delta
}
