package api

import (
	"crypto/tls"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"time"
)

type SysConfigResp struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// ==============================
// 模型检测（自己实现）
// ==============================
func checkModelExist() bool {
	// TODO: 按你实际逻辑
	return true
}

// ==============================
// GET /sys_config
// ==============================
func SysConfig(c *gin.Context) {

	// 线程1  检查状态
	// 线程2 下载依赖

}
func SysInit(c *gin.Context) {

	defaultData := map[string]interface{}{
		"init": true,
	}

	// 判断模型加载
	defaultData["init"] = checkModelExist()

	// ==============================
	// HTTP Client
	// ==============================
	client := &http.Client{
		Timeout: 60 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // verify=False
			},
		},
	}

	req, err := http.NewRequest(
		"GET",
		"https://www.xiacut.com/vapi/app/ai/client/sys_config",
		nil,
	)
	if err != nil {
		Err(c, err)
		return
	}

	// 自定义 Header（如有）
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		Err(c, err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		Err(c, err)
		return
	}

	// ==============================
	// JSON 解析
	// ==============================
	var jsonObj map[string]interface{}
	if err := json.Unmarshal(body, &jsonObj); err != nil {
		Err(c, err)
		return
	}

	// ==============================
	// 提取 version_info
	// ==============================
	versionInfo := map[string]interface{}{}

	data, ok := jsonObj["data"].(map[string]interface{})
	if ok {
		if configs, ok := data["sys_configs"].([]interface{}); ok {

			for _, item := range configs {
				cfg := item.(map[string]interface{})

				code := cfg["code"].(string)
				content := cfg["content"]

				versionInfo[code] = content
			}
		}
	}

	defaultData["version_info"] = versionInfo

	// ==============================
	// 返回
	// ==============================

	OK(c, defaultData)
}
