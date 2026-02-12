package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
	"xiacutai-server/internal/component/errs"
)

var DataDir string
var RegistryFile string
var SQLiteFile string
var StorageDir string
var JsonDir string

func InitDirs() {

	// 优先环境变量
	env := os.Getenv("DataDir")

	if env != "" {
		DataDir = env
	} else {
		DataDir = filepath.Join(GetExeDir(), "data")
	}

	RegistryFile = filepath.Join(DataDir, "models.json")
	SQLiteFile = filepath.Join(DataDir, "xiacutai.db")
	StorageDir = filepath.Join(DataDir, "storage")
	JsonDir = filepath.Join(DataDir, "json")

	// ===== 创建目录 =====
	mustMkdir(DataDir)
	mustMkdir(StorageDir)
	mustMkdir(JsonDir)
}
func mustMkdir(dir string) {
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		panic(fmt.Sprintf("failed to create dir %s: %v", dir, err))
	}
}

// 获取程序运行目录
func GetExeDir() string {

	exe, _ := os.Executable()
	exeDir := filepath.Dir(exe)

	// 2. IDE / go run 临时目录特征
	if strings.Contains(exeDir, "GoLand") ||
		strings.Contains(exeDir, "go-build") ||
		strings.Contains(exeDir, "Temp") {

		wd, _ := os.Getwd()
		return wd
	}

	// 3. 正式运行
	return exeDir
}

func CopyToStorage(src string) (string, error) {

	// 检查源文件
	if _, err := os.Stat(src); err != nil {
		return "", errs.New(fmt.Sprintf("file not exist: %v", err))
	}

	// 唯一文件名（避免并发冲突）
	ext := filepath.Ext(src)
	newName := fmt.Sprintf("%d_%d%s", time.Now().UnixNano(), os.Getpid(), ext)

	dst := filepath.Join(StorageDir, newName)

	// 打开源文件
	in, err := os.Open(src)
	if err != nil {
		return "", err
	}
	defer in.Close()

	// 重试创建（解决 Windows 占用）
	var out *os.File
	for i := 0; i < 5; i++ {
		out, err = os.Create(dst)
		if err == nil {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	if err != nil {
		return "", err
	}
	defer out.Close()

	// 拷贝
	if _, err = io.Copy(out, in); err != nil {
		return "", err
	}

	// 强制刷盘（关键）
	if err = out.Sync(); err != nil {
		return "", err
	}

	// 返回相对路径（非常关键）

	return dst, nil
}
