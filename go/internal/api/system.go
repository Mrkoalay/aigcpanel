package api

import (
	"crypto/tls"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"time"
	"xiacutai-server/internal/domain"
	"xiacutai-server/internal/service"
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
func SysConfig(c *gin.Context) {

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

// ==============================
// GET /sys_config
// ==============================
func SysInit(c *gin.Context) {
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
	modelInfos := map[string]interface{}{}

	data, ok := jsonObj["data"].(map[string]interface{})
	if ok {
		if configs, ok := data["sys_configs"].([]interface{}); ok {

			for _, item := range configs {
				cfg := item.(map[string]interface{})

				code := cfg["code"].(string)
				content := cfg["content"]

				modelInfos[code] = content
			}
		}
	}
	i := modelInfos["model_infos"].(string)

	bytei := []byte(i)
	var localModelRegistryModels []domain.LocalModelRegistryModel
	if err := json.Unmarshal(bytei, &localModelRegistryModels); err != nil {
		Err(c, err)
		return
	}

	// 这个接口可以重复调用，如果有任务进行中则需要返回进行中以及进度
	// 加锁以防止重复执行
	for _, localModel := range localModelRegistryModels {
		key := localModel.Key
		// 本地仓库如果没有
		//则根据URL下载到本地，并且把状态置为2
		// 调用添加模型接口把下载后的config 文件添加到DB
		localModelConfigInfo, err := service.Model.Get(key)
		if err != nil {
			Err(c, err)
		}
		// 如果本地有，状态是未下载依赖，则进入模型目录执行下载依赖命令，完成之后更新状态
		// 可以异步去下载依赖  因为时间会很久
	}

}
