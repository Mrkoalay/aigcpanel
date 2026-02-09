package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

func InitDirs() error {

	// 主数据目录
	if err := os.MkdirAll(DataDir, 0755); err != nil {
		return err
	}

	// storage 目录
	if err := os.MkdirAll(StorageDir, 0755); err != nil {
		return err
	}

	return nil
}
func CopyToStorage(src string) (string, error) {

	// 检查源文件
	if _, err := os.Stat(src); err != nil {
		return "", fmt.Errorf("file not exist: %v", err)
	}

	// 生成新文件名
	ext := filepath.Ext(src)
	newName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)

	dst := filepath.Join(StorageDir, newName)

	in, err := os.Open(src)
	if err != nil {
		return "", err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return "", err
	}
	defer out.Close()

	if _, err = io.Copy(out, in); err != nil {
		return "", err
	}

	return dst, nil
}
