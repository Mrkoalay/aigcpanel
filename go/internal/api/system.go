package api

import (
	"archive/zip"
	"bufio"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
	"xiacutai-server/internal/component/log"
	"xiacutai-server/internal/domain"
	"xiacutai-server/internal/service"
)

type SysConfigResp struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// ==============================
// 中文日志工具（适配：Info(msg string, fields ...zap.Field)）
// ==============================

func info(msg string, fields ...zap.Field) {
	log.Info("【系统初始化】"+msg, fields...)
}
func warn(msg string, fields ...zap.Field) {
	log.Warn("【系统初始化】"+msg, fields...)
}
func errlog(msg string, fields ...zap.Field) {
	log.Error("【系统初始化】"+msg, fields...)
}

// defer step("xxx", zap.String("k","v"))()
func step(name string, fields ...zap.Field) func() {
	start := time.Now()
	info("开始："+name, fields...)
	return func() {
		fs := append(fields, zap.Duration("耗时", time.Since(start)))
		info("完成："+name, fs...)
	}
}

func previewBytes(b []byte, n int) string {
	if len(b) <= n {
		return string(b)
	}
	return string(b[:n]) + "...(已截断)"
}

// ==============================
// 模型检测（自己实现）
// ==============================
func checkModelExist() bool {
	// TODO: 按你实际逻辑
	return true
}

// ==============================
// 复用 HTTP Client（避免每次 new Transport）
// ==============================
var sysHTTPClient = &http.Client{
	Timeout: 60 * time.Second,
	Transport: &http.Transport{
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
		MaxIdleConns:        50,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	},
}

// ==============================
// sys_config 拉取 + 解析工具
// ==============================

type sysConfigItem struct {
	Code    string          `json:"code"`
	Content json.RawMessage `json:"content"`
}

type sysConfigData struct {
	SysConfigs []sysConfigItem `json:"sys_configs"`
}

type sysConfigRespTyped struct {
	Code    int           `json:"code"`
	Message string        `json:"message"`
	Data    sysConfigData `json:"data"`
}

// 拉取 sys_config
func fetchSysConfig(ctx context.Context) (*sysConfigRespTyped, []byte, error) {
	defer step("拉取远端 sys_config")()

	url := "https://www.xiacut.com/vapi/app/ai/client/sys_config"
	info("准备请求 sys_config", zap.String("url", url))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		errlog("创建请求失败", zap.Error(err))
		return nil, nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := sysHTTPClient.Do(req)
	if err != nil {
		errlog("请求 sys_config 失败", zap.Error(err))
		return nil, nil, err
	}
	defer resp.Body.Close()

	info("收到 sys_config 响应", zap.Int("status_code", resp.StatusCode), zap.String("status", resp.Status), zap.Int64("content_length", resp.ContentLength))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 8*1024))
		errlog("sys_config 响应非 2xx", zap.String("status", resp.Status), zap.String("body_preview", string(b)))
		return nil, nil, errors.New("sys_config http " + resp.Status + ": " + string(b))
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 4*1024*1024))
	if err != nil {
		errlog("读取 sys_config body 失败", zap.Error(err))
		return nil, nil, err
	}
	info("读取 sys_config body 完成", zap.Int("body_size", len(body)))

	var obj sysConfigRespTyped
	if err := json.Unmarshal(body, &obj); err != nil {
		errlog("解析 sys_config JSON 失败", zap.Error(err))
		return nil, body, err
	}

	info("解析 sys_config 成功", zap.Int("code", obj.Code), zap.String("message", obj.Message), zap.Int("sys_configs", len(obj.Data.SysConfigs)))
	return &obj, body, nil
}

// 将 sys_configs 转成 map[code]content(尽量还原成 string/json)
func buildSysConfigMap(items []sysConfigItem) map[string]interface{} {
	defer step("解析 sys_configs", zap.Int("items", len(items)))()

	out := make(map[string]interface{}, len(items))
	for _, it := range items {
		if it.Code == "" {
			continue
		}

		var s string
		if err := json.Unmarshal(it.Content, &s); err == nil {
			out[it.Code] = s
			info("解析 sys_config 项", zap.String("code", it.Code), zap.String("content_type", "string"), zap.Int("len", len(s)))
			continue
		}

		var anyVal interface{}
		if err := json.Unmarshal(it.Content, &anyVal); err == nil {
			out[it.Code] = anyVal
			info("解析 sys_config 项", zap.String("code", it.Code), zap.String("content_type", "json"))
			continue
		}

		out[it.Code] = it.Content
		warn("解析 sys_config 项为 RawMessage", zap.String("code", it.Code), zap.Int("raw_len", len(it.Content)))
	}
	return out
}

// 专门解析 model_infos（兼容：content 是 JSON 字符串 或 直接 JSON）
func parseModelInfos(modelInfosVal interface{}) ([]domain.LocalModelRegistryModel, error) {
	defer step("解析 model_infos")()

	if modelInfosVal == nil {
		info("model_infos 为空，跳过")
		return nil, nil
	}

	var payload []byte
	switch v := modelInfosVal.(type) {
	case string:
		payload = []byte(v)
		info("model_infos 类型=string", zap.Int("bytes", len(payload)))
	case []interface{}, map[string]interface{}:
		b, err := json.Marshal(v)
		if err != nil {
			errlog("model_infos 对象转 JSON 失败", zap.Error(err))
			return nil, err
		}
		payload = b
		info("model_infos 类型=object", zap.Int("bytes", len(payload)))
	case json.RawMessage:
		payload = []byte(v)
		info("model_infos 类型=raw", zap.Int("bytes", len(payload)))
	case []byte:
		payload = v
		info("model_infos 类型=bytes", zap.Int("bytes", len(payload)))
	default:
		b, err := json.Marshal(v)
		if err != nil {
			errlog("model_infos fallback Marshal 失败", zap.Error(err), zap.String("type", fmt.Sprintf("%T", v)))
			return nil, err
		}
		payload = b
		info("model_infos 类型=fallback", zap.String("type", fmt.Sprintf("%T", v)), zap.Int("bytes", len(payload)))
	}

	var models []domain.LocalModelRegistryModel
	if err := json.Unmarshal(payload, &models); err != nil {
		errlog("解析 registryModels 失败", zap.Error(err), zap.String("payload_preview", previewBytes(payload, 512)))
		return nil, err
	}

	info("解析 registryModels 成功", zap.Int("models", len(models)))
	return models, nil
}

// ==============================
// GET /sys_config
// ==============================
func SysConfig(c *gin.Context) {
	defer step("SysConfig 接口处理")()

	defaultData := map[string]interface{}{
		"init": true,
	}
	defaultData["init"] = checkModelExist()

	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()

	obj, _, err := fetchSysConfig(ctx)
	if err != nil {
		Err(c, err)
		return
	}

	versionInfo := buildSysConfigMap(obj.Data.SysConfigs)
	defaultData["version_info"] = versionInfo

	OK(c, defaultData)
}

// ==============================
// SysInit：可重复调用 + 进度返回 + 异步初始化
// ==============================

type initState string

const (
	initIdle    initState = "idle"
	initRunning initState = "running"
	initDone    initState = "done"
	initFailed  initState = "failed"
)

type modelProgress struct {
	Key      string    `json:"key"`
	Status   int       `json:"status"`   // 2下载中/4依赖中/5完成/-1失败
	Progress int       `json:"progress"` // 0~100
	Message  string    `json:"message"`
	Error    string    `json:"error,omitempty"`
	Updated  time.Time `json:"updated"`
}

type sysInitTask struct {
	Status    initState                 `json:"status"`
	StartedAt time.Time                 `json:"started_at"`
	UpdatedAt time.Time                 `json:"updated_at"`
	Models    map[string]*modelProgress `json:"models"`
	Error     string                    `json:"error,omitempty"`
}

var (
	initMu   sync.Mutex
	initTask *sysInitTask
)

// GET /sys_init
func SysInit(c *gin.Context) {
	defer step("SysInit 接口处理")()

	// 1) 若已有任务在跑：直接返回进度（接口可重复调用）
	initMu.Lock()
	if initTask != nil && initTask.Status == initRunning {
		snap := snapshotInitLocked(initTask)
		initMu.Unlock()

		info("已有任务在执行，返回进度", zap.Int("models", len(snap.Models)))
		OK(c, gin.H{"running": true, "progress": snap})
		return
	}
	initMu.Unlock()

	// 2) 拉 sys_config
	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()

	obj, _, err := fetchSysConfig(ctx)
	if err != nil {
		errlog("拉取 sys_config 失败", zap.Error(err))
		Err(c, err)
		return
	}

	// 3) 提取 model_infos
	m := buildSysConfigMap(obj.Data.SysConfigs)
	registryModels, err := parseModelInfos(m["model_infos"])
	if err != nil {
		errlog("解析 model_infos 失败", zap.Error(err))
		Err(c, err)
		return
	}
	info("准备初始化模型", zap.Int("count", len(registryModels)))

	// 4) 初始化任务并启动异步
	task := &sysInitTask{
		Status:    initRunning,
		StartedAt: time.Now(),
		UpdatedAt: time.Now(),
		Models:    map[string]*modelProgress{},
	}

	// ✅ 关键修复：创建 task 时先查 DB，把已就绪的直接标成“就绪”，不要“排队中”
	for _, rm := range registryModels {
		key := strings.TrimSpace(rm.Key)
		if key == "" {
			continue
		}

		st := 0
		pg := 0
		msg := "排队中"

		local, dbErr := service.Model.GetByDB(key)
		if dbErr == nil && local != nil && local.Status == "5" {
			st = 5
			pg = 100
			msg = "就绪"
		}

		task.Models[key] = &modelProgress{
			Key:      key,
			Status:   st,
			Progress: pg,
			Message:  msg,
			Updated:  time.Now(),
		}
	}

	initMu.Lock()
	if initTask != nil && initTask.Status == initRunning {
		snap := snapshotInitLocked(initTask)
		initMu.Unlock()
		info("并发触发：已存在任务，返回进度", zap.Int("models", len(snap.Models)))
		OK(c, gin.H{"running": true, "progress": snap})
		return
	}
	initTask = task
	initMu.Unlock()

	info("初始化任务已创建，开始异步执行", zap.Int("models", len(task.Models)))
	go runInit(registryModels)

	OK(c, gin.H{"running": true, "progress": snapshotInit(task)})
}

func runInit(models []domain.LocalModelRegistryModel) {
	defer step("runInit 模型初始化总流程", zap.Int("models", len(models)))()

	for i, rm := range models {
		key := strings.TrimSpace(rm.Key)
		if key == "" {
			warn("跳过模型：key 为空", zap.Int("index", i))
			continue
		}

		// 查 DB
		t0 := time.Now()
		localModelConfigInfo, err := service.Model.GetByDB(key)
		info("DB 查询完成",
			zap.String("key", key),
			zap.Duration("耗时", time.Since(t0)),
			zap.Bool("为空", localModelConfigInfo == nil),
			zap.Error(err),
		)

		info("开始处理模型", zap.String("key", key), zap.Int("index", i+1), zap.Int("total", len(models)))

		// ✅ 关键修复：已就绪的模型，立即更新为“就绪”，不要保留“排队中”
		if localModelConfigInfo != nil && localModelConfigInfo.Status == "5" {
			updateModel(key, 5, 100, "就绪", "")
			continue
		}

		// ✅ 关键修复：避免 nil panic + 语义更清晰
		exists := localModelConfigInfo != nil && strings.TrimSpace(localModelConfigInfo.Key) != ""
		if !exists {
			updateModel(key, 2, 20, "未发现模型，开始下载", "")
			info("开始下载并解压", zap.String("key", key), zap.String("url", rm.URL))

			t1 := time.Now()
			err = downloadModelPlaceholder(rm)
			info("下载并解压完成", zap.String("key", key), zap.Duration("耗时", time.Since(t1)), zap.Error(err))

			if err != nil {
				updateModel(key, -1, 100, "下载/解压/添加失败", err.Error())
				continue
			}
			updateModel(key, 3, 60, "下载完成", "")
		} else {
			info("模型已存在，跳过下载", zap.String("key", key))
		}

		localModelConfigInfo, err = service.Model.GetByDB(key)
		if err != nil {
			updateModel(key, -1, 100, "DB 查询失败", err.Error())
			continue
		}
		if localModelConfigInfo == nil {
			updateModel(key, -1, 100, "模型信息缺失", "db returned nil after download")
			continue
		}

		if localModelConfigInfo.Status == "3" {
			// 安装依赖（占位）
			updateModel(key, 4, 75, "安装依赖中", "")
			info("开始安装依赖", zap.String("key", key))

			t2 := time.Now()
			err = installDepsPlaceholder(key)
			info("安装依赖完成", zap.String("key", key), zap.Duration("耗时", time.Since(t2)), zap.Error(err))
			if err != nil {
				updateModel(key, -1, 100, "依赖安装失败", err.Error())
				continue
			}

			updateModel(key, 5, 100, "就绪", "")
			info("模型就绪", zap.String("key", key))
		} else if localModelConfigInfo.Status == "5" {
			// 双保险：可能 DB 状态被其他流程更新到 5
			updateModel(key, 5, 100, "就绪", "")
		}
	}

	// 任务收尾
	initMu.Lock()
	defer initMu.Unlock()

	if initTask == nil {
		return
	}

	fail := false
	for _, p := range initTask.Models {
		if p != nil && p.Status == -1 {
			fail = true
			break
		}
	}
	initTask.UpdatedAt = time.Now()
	if fail {
		initTask.Status = initFailed
		initTask.Error = "存在模型初始化失败"
		errlog("初始化任务结束：失败")
	} else {
		initTask.Status = initDone
		info("初始化任务结束：成功")
	}
}

// 你可以把 models 根目录做成配置项；这里先用相对路径示例
var modelsRootDir = "models"

// downloadModelPlaceholder：下载 zip -> 解压 -> 找 config -> ModelAdd
func downloadModelPlaceholder(rm domain.LocalModelRegistryModel) error {
	defer step("下载模型完整流程", zap.String("key", rm.Key))()

	key := strings.TrimSpace(rm.Key)
	if key == "" {
		errlog("模型 key 为空，无法下载")
		return errors.New("empty model key")
	}

	url := rm.URL
	if url == "" {
		errlog("缺少下载地址", zap.String("key", key))
		return fmt.Errorf("model %s missing download url", key)
	}

	info("下载参数", zap.String("key", key), zap.String("url", url))

	// 下载 zip
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	t0 := time.Now()
	tmpZipPath, err := downloadToTempZip(ctx, sysHTTPClient, url)
	info("下载压缩包完成",
		zap.String("key", key),
		zap.String("tmp_zip", tmpZipPath),
		zap.Duration("耗时", time.Since(t0)),
		zap.Error(err),
	)
	if err != nil {
		return err
	}
	defer func() {
		_ = os.Remove(tmpZipPath)
		info("清理临时压缩包", zap.String("tmp_zip", tmpZipPath), zap.String("key", key))
	}()

	// 解压
	info("解压目标目录", zap.String("dest_dir", modelsRootDir), zap.String("key", key))

	if err := os.MkdirAll(modelsRootDir, 0o755); err != nil {
		errlog("创建目录失败", zap.String("dest_dir", modelsRootDir), zap.Error(err))
		return err
	}

	t1 := time.Now()
	files, err := unzipSafeWithCount(tmpZipPath, modelsRootDir)
	info("解压完成",
		zap.String("key", key),
		zap.Int("files", files),
		zap.Duration("耗时", time.Since(t1)),
		zap.Error(err),
	)
	if err != nil {
		return err
	}

	// 找 config
	t2 := time.Now()
	dirName := getDirByKey(key)
	cfgPath, err := findConfigFile(filepath.Join(modelsRootDir, dirName))
	info("查找 config 完成",
		zap.String("key", key),
		zap.String("cfg_path", cfgPath),
		zap.Duration("耗时", time.Since(t2)),
		zap.Error(err),
	)
	if err != nil {
		return fmt.Errorf("model %s: %w", key, err)
	}

	// ModelAdd
	info("开始调用 ModelAdd", zap.String("key", key), zap.String("cfg_path", cfgPath))
	t3 := time.Now()
	out, err := service.Model.ModelAdd(cfgPath)
	info("ModelAdd 完成",
		zap.String("key", key),
		zap.String("cfg_path", cfgPath),
		zap.String("out_type", fmt.Sprintf("%T", out)),
		zap.Duration("耗时", time.Since(t3)),
		zap.Error(err),
	)
	if err != nil {
		return fmt.Errorf("model %s ModelAdd failed: %w", key, err)
	}

	return nil
}

func getDirByKey(key string) string {
	keySplits := strings.Split(key, "|")
	dirName := keySplits[0] + "-win-x86-v" + keySplits[1]
	return dirName
}

// 下载 URL 到一个临时 zip 文件，返回临时文件路径
func downloadToTempZip(ctx context.Context, client *http.Client, url string) (string, error) {
	defer step("下载压缩包", zap.String("url", url))()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		errlog("创建下载请求失败", zap.Error(err))
		return "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		errlog("下载请求失败", zap.Error(err))
		return "", err
	}
	defer resp.Body.Close()

	info("下载响应", zap.String("status", resp.Status), zap.Int("status_code", resp.StatusCode), zap.Int64("content_length", resp.ContentLength))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 8*1024))
		errlog("下载失败（非 2xx）", zap.String("status", resp.Status), zap.String("body_preview", string(b)))
		return "", fmt.Errorf("download failed: %s: %s", resp.Status, string(b))
	}

	const maxBytes = int64(2 << 30) // 2GB
	body := io.LimitReader(resp.Body, maxBytes+1)

	f, err := os.CreateTemp("", "model_*.zip")
	if err != nil {
		errlog("创建临时文件失败", zap.Error(err))
		return "", err
	}
	defer func() {
		_ = f.Close()
		if err != nil {
			_ = os.Remove(f.Name())
		}
	}()

	info("临时 zip 路径", zap.String("tmp_zip", f.Name()))

	n, err := io.Copy(f, body)
	if err != nil {
		errlog("写入临时 zip 失败", zap.Error(err))
		return "", err
	}
	if n > maxBytes {
		errlog("压缩包过大", zap.Int64("bytes", n), zap.Int64("max", maxBytes))
		return "", fmt.Errorf("zip too large: %d > %d", n, maxBytes)
	}

	info("写入 zip 完成", zap.Int64("bytes", n))
	return f.Name(), nil
}

// 安全解压：防止 Zip Slip（.. 路径穿越），并统计写入文件数
func unzipSafeWithCount(zipPath, destDir string) (int, error) {
	defer step("解压压缩包", zap.String("zip", zipPath), zap.String("dest", destDir))()

	r, err := zip.OpenReader(zipPath)
	if err != nil {
		errlog("打开 zip 失败", zap.Error(err))
		return 0, err
	}
	defer r.Close()

	destDirClean := filepath.Clean(destDir) + string(os.PathSeparator)
	filesWritten := 0

	for _, f := range r.File {
		name := f.Name

		if strings.HasPrefix(name, "__MACOSX/") {
			continue
		}

		targetPath := filepath.Join(destDir, filepath.FromSlash(name))
		targetClean := filepath.Clean(targetPath)

		if !strings.HasPrefix(targetClean+string(os.PathSeparator), destDirClean) &&
			targetClean != strings.TrimRight(destDirClean, string(os.PathSeparator)) {
			errlog("检测到非法路径（Zip Slip）", zap.String("name", name))
			return filesWritten, fmt.Errorf("illegal zip path: %s", name)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(targetClean, 0o755); err != nil {
				errlog("创建目录失败", zap.String("dir", targetClean), zap.Error(err))
				return filesWritten, err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(targetClean), 0o755); err != nil {
			errlog("创建父目录失败", zap.String("dir", filepath.Dir(targetClean)), zap.Error(err))
			return filesWritten, err
		}

		rc, err := f.Open()
		if err != nil {
			errlog("打开 zip 内文件失败", zap.String("name", name), zap.Error(err))
			return filesWritten, err
		}

		out, err := os.OpenFile(targetClean, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
		if err != nil {
			_ = rc.Close()
			errlog("创建文件失败", zap.String("file", targetClean), zap.Error(err))
			return filesWritten, err
		}

		n, cpErr := io.Copy(out, rc)
		_ = out.Close()
		_ = rc.Close()

		if cpErr != nil {
			errlog("写文件失败", zap.String("file", targetClean), zap.Error(cpErr))
			return filesWritten, cpErr
		}

		filesWritten++
		if filesWritten <= 5 {
			info("解压文件", zap.String("file", targetClean), zap.Int64("bytes", n))
		}
	}

	info("解压结束", zap.Int("files", filesWritten))
	return filesWritten, nil
}

// 在解压目录中寻找 config 文件
func findConfigFile(root string) (string, error) {
	defer step("查找 config 文件", zap.String("root", root))()

	var candidates []string

	entries, err := os.ReadDir(root)
	if err != nil {
		errlog("读取目录失败", zap.String("root", root), zap.Error(err))
		return "", err
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := strings.ToLower(e.Name())
		if strings.HasSuffix(name, ".json") {
			p := filepath.Join(root, e.Name())
			candidates = append(candidates, p)
			info("发现候选 config（根目录）", zap.String("path", p))
		}
	}

	if len(candidates) == 0 {
		var found string
		err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			name := strings.ToLower(d.Name())
			if strings.HasSuffix(name, ".json") {
				found = path
				info("发现候选 config（递归）", zap.String("path", found))
				return io.EOF
			}
			return nil
		})
		if err != nil && err != io.EOF {
			errlog("递归查找失败", zap.Error(err))
			return "", err
		}
		if found == "" {
			errlog("未找到任何 config json", zap.String("root", root))
			return "", errors.New("config json not found")
		}
		info("选定 config", zap.String("path", found))
		return found, nil
	}

	best := candidates[0]
	bestScore := scoreConfigPath(best)
	info("初始最优 config", zap.String("path", best), zap.Int("score", bestScore))

	for _, p := range candidates[1:] {
		s := scoreConfigPath(p)
		info("候选 config", zap.String("path", p), zap.Int("score", s))
		if s > bestScore {
			best = p
			bestScore = s
		}
	}

	info("最终选定 config", zap.String("path", best), zap.Int("score", bestScore))
	return best, nil
}

func scoreConfigPath(p string) int {
	name := strings.ToLower(filepath.Base(p))
	score := 0
	if strings.Contains(name, "musetalk") {
		score += 50
	}
	if strings.Contains(name, "cosyvoice") {
		score += 50
	}
	if strings.Contains(name, "model") {
		score += 10
	}
	if strings.Contains(name, "config") {
		score += 10
	}
	return score
}

// ==============================
// 进度更新/快照
// ==============================

func updateModel(key string, status, progress int, msg string, errMsg string) {
	initMu.Lock()
	defer initMu.Unlock()

	if initTask == nil {
		return
	}
	p, ok := initTask.Models[key]
	if !ok {
		p = &modelProgress{Key: key}
		initTask.Models[key] = p
	}
	p.Status = status
	p.Progress = progress
	p.Message = msg
	p.Updated = time.Now()
	if errMsg != "" {
		p.Error = errMsg
	}
	initTask.UpdatedAt = time.Now()
	service.Model.ModelUpdateStatus(key, status)
	if errMsg != "" {
		errlog("进度更新", zap.String("key", key), zap.Int("status", status), zap.Int("progress", progress), zap.String("msg", msg), zap.String("err", errMsg))
	} else {
		info("进度更新", zap.String("key", key), zap.Int("status", status), zap.Int("progress", progress), zap.String("msg", msg))
	}
}

func snapshotInit(task *sysInitTask) *sysInitTask {
	initMu.Lock()
	defer initMu.Unlock()
	return snapshotInitLocked(task)
}

func snapshotInitLocked(task *sysInitTask) *sysInitTask {
	if task == nil {
		return nil
	}
	cp := *task
	cp.Models = map[string]*modelProgress{}
	for k, v := range task.Models {
		if v == nil {
			continue
		}
		p := *v
		cp.Models[k] = &p
	}
	return &cp
}

// ==============================
// 占位：依赖安装（你后续替换成真实逻辑）
// ==============================
// installDepsPlaceholder：进入模型目录执行 launch.bat
func installDepsPlaceholder(key string) error {
	defer step("安装依赖：执行 launch.bat", zap.String("key", key))()

	dirByKey := getDirByKey(key)
	modelDir := filepath.Join(modelsRootDir, dirByKey)
	batPath := filepath.Join(modelDir, "launch.bat")

	info("准备安装依赖（执行脚本）",
		zap.String("key", key),
		zap.String("model_dir", modelDir),
		zap.String("bat", batPath),
	)

	// 基本检查
	if st, err := os.Stat(modelDir); err != nil || !st.IsDir() {
		errlog("模型目录不存在或不可用",
			zap.String("key", key),
			zap.String("model_dir", modelDir),
			zap.Error(err),
		)
		if err == nil {
			err = errors.New("not a directory")
		}
		return fmt.Errorf("model dir invalid: %w", err)
	}
	if _, err := os.Stat(batPath); err != nil {
		errlog("launch.bat 不存在",
			zap.String("key", key),
			zap.String("bat", batPath),
			zap.Error(err),
		)
		return fmt.Errorf("launch.bat not found: %w", err)
	}

	// 超时控制（你要更久就改这里）
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Minute)
	defer cancel()

	start := time.Now()
	err := runBatWithLiveLogs(ctx, key, modelDir, "launch.bat")
	costDur := time.Since(start)

	if err != nil {
		errlog("安装依赖失败（launch.bat 执行失败）",
			zap.String("key", key),
			zap.Duration("耗时", costDur),
			zap.Error(err),
		)
		return err
	}

	info("安装依赖成功（launch.bat 执行完成）",
		zap.String("key", key),
		zap.Duration("耗时", costDur),
	)
	return nil
}

func runBatWithLiveLogs(ctx context.Context, key string, dir string, batName string) error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("当前系统不是 Windows，无法执行 bat：%s", batName)
	}

	// cmd.exe /C launch.bat
	cmd := exec.CommandContext(ctx, "cmd.exe", "/C", batName)
	cmd.Dir = dir
	cmd.Env = os.Environ()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("StdoutPipe 失败: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("StderrPipe 失败: %w", err)
	}

	info("开始执行脚本",
		zap.String("key", key),
		zap.String("dir", dir),
		zap.String("cmd", "cmd.exe /C "+batName),
	)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动脚本失败: %w", err)
	}

	// 并发读取 stdout/stderr，逐行打日志
	errCh := make(chan error, 2)
	go streamLinesToLog(key, "stdout", stdout, errCh)
	go streamLinesToLog(key, "stderr", stderr, errCh)

	// 等命令结束
	waitErr := cmd.Wait()

	// 等待读完（不阻塞太久）
	select {
	case <-errCh:
	default:
	}

	// 超时/取消
	if ctx.Err() == context.DeadlineExceeded {
		return fmt.Errorf("脚本执行超时：%w", ctx.Err())
	}
	if ctx.Err() == context.Canceled {
		return fmt.Errorf("脚本被取消：%w", ctx.Err())
	}

	if waitErr != nil {
		// 尝试拿 exit code
		exitCode := -1
		if ee, ok := waitErr.(*exec.ExitError); ok {
			exitCode = ee.ExitCode()
		}
		return fmt.Errorf("脚本执行失败 exitCode=%d err=%w", exitCode, waitErr)
	}

	info("脚本执行完成",
		zap.String("key", key),
		zap.String("bat", batName),
	)
	return nil
}

func streamLinesToLog(key string, stream string, r io.Reader, errCh chan<- error) {
	// bufio.Scanner 默认 token 太小，bat 输出长行可能炸，扩大一下
	sc := bufio.NewScanner(r)
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 4*1024*1024) // 单行最大 4MB

	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		// 实时日志：每行一条
		info("脚本输出",
			zap.String("key", key),
			zap.String("流", stream),
			zap.String("内容", line),
		)
	}
	if err := sc.Err(); err != nil {
		// Scanner 读取出错也记录一下，但不一定要让主流程失败
		warn("读取脚本输出失败",
			zap.String("key", key),
			zap.String("流", stream),
			zap.Error(err),
		)
		select {
		case errCh <- err:
		default:
		}
		return
	}
	select {
	case errCh <- nil:
	default:
	}
}
