package api

import (
    "encoding/json"
    "strings"
    "time"
    "xiacutai-server/internal/component/errs"
    "xiacutai-server/internal/component/sqllite"
    "xiacutai-server/internal/domain"
    "xiacutai-server/internal/service"

    "github.com/gin-gonic/gin"
)

type taskCreateRequest struct {
    Text      string                 `json:"text"`
    Type      string                 `json:"type"`
    ServerKey string                 `json:"serverKey"`
    Param     map[string]any         `json:"param"`
    Extra     map[string]interface{} `json:"-"`
}
type taskOperateRequest struct {
    ID int64 `json:"id"`
}
type taskUpdateRequest struct {
    ID    int64  `json:"id"`
    Title string `json:"title"`
}
type soundGenerateRequest struct {
    Text           string         `json:"text"`
    Type           string         `json:"type"`
    ServerName     string         `json:"serverName"`
    ServerTitle    string         `json:"serverTitle"`
    ServerVersion  string         `json:"serverVersion"`
    TtsServerKey   string         `json:"ttsServerKey"`
    TtsParam       map[string]any `json:"ttsParam"`
    CloneServerKey string         `json:"cloneServerKey"`
    CloneParam     map[string]any `json:"cloneParam"`
    PromptID       int64          `json:"promptId"`
    PromptTitle    string         `json:"promptTitle"`
    PromptURL      string         `json:"promptUrl"`
    PromptText     string         `json:"promptText"`
}

var TypeBizMap = map[string]string{
    "soundTts": "SoundGenerate",
}

func DataTaskCreate(ctx *gin.Context) {
    var req taskCreateRequest
    if err := ctx.ShouldBindJSON(&req); err != nil {
        Err(ctx, err)
        return
    }
    if strings.TrimSpace(req.Text) == "" {
        Err(ctx, errs.ParamError)
        return
    }
    serverKey := req.ServerKey
    model, err := service.Model.Get(serverKey)

    typeStr := req.Type
    modelConfig := map[string]any{}

    switch typeStr {
    case "soundTts":
        modelConfig = map[string]any{
            "type":         typeStr,
            "ttsServerKey": serverKey,
            "ttsParam":     req.Param,
            "text":         req.Text,
        }
    case "soundClone":
        //result, err = server.SoundClone(*functionData)
    case "videoGen":
        //result, err = server.VideoGen(*functionData)
    case "asr":
        //result, err = server.Asr(*functionData)
    default:
        Err(ctx, errs.New("不支持的功能"))
        return
    }

    modelConfigRaw, err := json.Marshal(modelConfig)
    if err != nil {
        Err(ctx, err)
        return
    }
    paramRaw, err := json.Marshal(map[string]any{})
    if err != nil {
        Err(ctx, err)
        return
    }

    if err != nil {
        Err(ctx, err)
        return
    }
    task := domain.DataTaskModel{
        Biz:           TypeBizMap[typeStr],
        Title:         req.Text,
        Status:        "queue",
        ServerName:    model.Name,
        ServerTitle:   model.Title,
        ServerVersion: model.Version,
        Param:         string(paramRaw),
        ModelConfig:   string(modelConfigRaw),
        JobResult:     "{}",
        Result:        "{}",
        Type:          1,
    }
    created, err := service.DataTask.CreateTask(task)
    if err != nil {
        Err(ctx, err)
        return
    }
    OK(ctx, gin.H{
        "data": created,
    })
}

type dataTaskListRequest struct {
    Biz  string `form:"biz"`
    Page int    `form:"page"`
    Size int    `form:"size"`
}

func DataTaskList(ctx *gin.Context) {
    var req dataTaskListRequest
    if err := ctx.ShouldBindJSON(&req); err != nil {
        Err(ctx, err)
        return
    }

    list, err := service.DataTask.ListTasks(sqllite.TaskFilters{
        Biz:  req.Biz,
        Page: req.Page,
        Size: req.Size,
    })
    if err != nil {
        Err(ctx, err)
        return
    }

    OK(ctx, gin.H{
        "data": list,
    })
}

func DataTaskCancel(ctx *gin.Context) {
    var req taskOperateRequest
    if err := ctx.ShouldBindJSON(&req); err != nil {
        Err(ctx, err)
        return
    }
    if req.ID <= 0 {
        Err(ctx, errs.ParamError)
        return
    }
    current, err := service.DataTask.GetTask(req.ID)
    if err != nil {
        Err(ctx, err)
        return
    }
    switch current.Status {
    case domain.TaskStatusQueue, domain.TaskStatusWait:
    case domain.TaskStatusRunning:
        if err := service.CancelEasyServerTask(req.ID); err != nil {
            Err(ctx, err)
            return
        }
    default:
        Err(ctx, errs.New("任务状态不允许取消"))
        return
    }
    task, err := service.DataTask.UpdateTask(req.ID, map[string]any{
        "status":    domain.TaskStatusFail,
        "statusMsg": "cancelled",
        "endTime":   time.Now().UnixMilli(),
    })
    if err != nil {
        Err(ctx, err)
        return
    }
    OK(ctx, gin.H{
        "data": task,
    })
}

func DataTaskContinue(ctx *gin.Context) {
    var req taskOperateRequest
    if err := ctx.ShouldBindJSON(&req); err != nil {
        Err(ctx, err)
        return
    }
    if req.ID <= 0 {
        Err(ctx, errs.ParamError)
        return
    }

    current, err := service.DataTask.GetTask(req.ID)
    if err != nil {
        Err(ctx, err)
        return
    }
    if current.Status == domain.TaskStatusFail {
        task, err := service.DataTask.UpdateTask(req.ID, map[string]any{
            "status":    domain.TaskStatusQueue,
            "statusMsg": "",
        })
        if err != nil {
            Err(ctx, err)
            return
        }
        OK(ctx, gin.H{
            "data": task,
        })
        return
    }

    OK(ctx, gin.H{})
}

func DataTaskDelete(ctx *gin.Context) {
    var req taskOperateRequest
    if err := ctx.ShouldBindJSON(&req); err != nil {
        Err(ctx, err)
        return
    }
    if req.ID <= 0 {
        Err(ctx, errs.ParamError)
        return
    }
    current, err := service.DataTask.GetTask(req.ID)
    if current.Status == domain.TaskStatusRunning {
        Err(ctx, errs.New("不可删除运行中任务"))
        return
    }
    if err != nil {
        Err(ctx, err)
        return
    }

    if err := service.DataTask.DeleteTask(req.ID); err != nil {
        Err(ctx, err)
        return
    }
    OK(ctx, gin.H{
        "data": req.ID,
    })
}
func DataTaskUpdate(ctx *gin.Context) {
    var req taskUpdateRequest
    if err := ctx.ShouldBindJSON(&req); err != nil {
        Err(ctx, err)
        return
    }
    if req.ID <= 0 {
        Err(ctx, errs.ParamError)
        return
    }

    title := req.Title
    updateMap := map[string]any{
        "title": title,
    }
    task, err := service.DataTask.UpdateTask(req.ID, updateMap)

    if err != nil {
        Err(ctx, err)
        return
    }

    OK(ctx, gin.H{
        "data": task,
    })
}
