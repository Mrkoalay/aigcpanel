package service

import (
	"fmt"
	"strings"
	"time"
	"xiacutai-server/internal/component/errs"
	"xiacutai-server/internal/component/modelcall/easyserver"
	"xiacutai-server/internal/domain"
)

func RecognizeSoundPromptText(audioPath string) (string, error) {
	if strings.TrimSpace(audioPath) == "" {
		return "", errs.ParamError
	}

	serverKey, err := findAsrServerKey()
	if err != nil {
		return "", err
	}

	server, err := startEasyServerByKey(serverKey)
	if err != nil {
		return "", err
	}
	defer server.Stop()

	result, err := server.Asr(easyserver.ServerFunctionDataType{
		ID:     fmt.Sprintf("storage-sound-asr-%d", time.Now().UnixMilli()),
		Param:  map[string]interface{}{},
		Result: map[string]interface{}{},
		Audio:  audioPath,
	})
	if err != nil {
		return "", err
	}

	resultData, err := extractResultData(result)
	if err != nil {
		return "", err
	}

	promptText := extractAsrPromptText(resultData)
	if promptText == "" {
		return "", errs.New("asr prompt text empty")
	}
	return promptText, nil
}

func findAsrServerKey() (string, error) {
	models, err := Model.ModelList(domain.FunctionSoundAsr)
	if err != nil {
		return "", err
	}
	if len(models) == 0 {
		return "", errs.New("未找到可用的语音识别模型")
	}

	model := models[0]
	if strings.TrimSpace(model.Key) != "" {
		return model.Key, nil
	}
	if strings.TrimSpace(model.Name) == "" || strings.TrimSpace(model.Version) == "" {
		return "", errs.New("语音识别模型配置不完整")
	}
	return model.Name + "|" + model.Version, nil
}

func extractAsrPromptText(resultData map[string]any) string {
	if text := strings.TrimSpace(asString(resultData["text"])); text != "" {
		return text
	}

	records, err := parseAsrRecords(resultData)
	if err != nil {
		return ""
	}

	var builder strings.Builder
	for _, record := range records {
		text := strings.TrimSpace(record.Text)
		if text == "" {
			continue
		}
		builder.WriteString(text)
	}
	return strings.TrimSpace(builder.String())
}
