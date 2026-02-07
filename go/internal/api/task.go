package api

import (
	"aigcpanel/go/internal/component/errs"
	"aigcpanel/go/internal/component/sqllite"
	"net/http"
	"strconv"
	"strings"

	"aigcpanel/go/internal/domain"
	"aigcpanel/go/internal/service"
	"github.com/gin-gonic/gin"
)

type taskPayload struct {
	ID            int64  `json:"id"`
	Biz           string `json:"biz"`
	Type          int    `json:"type"`
	Title         string `json:"title"`
	Status        string `json:"status"`
	StatusMsg     string `json:"statusMsg"`
	StartTime     int64  `json:"startTime"`
	EndTime       int64  `json:"endTime"`
	ServerName    string `json:"serverName"`
	ServerTitle   string `json:"serverTitle"`
	ServerVersion string `json:"serverVersion"`
	Param         string `json:"param"`
	JobResult     string `json:"jobResult"`
	ModelConfig   string `json:"modelConfig"`
	Result        string `json:"result"`
	CreatedAt     int64  `json:"createdAt"`
	UpdatedAt     int64  `json:"updatedAt"`
}

func TaskList(ctx *gin.Context) {
	statusQuery := strings.TrimSpace(ctx.Query("status"))
	var statusList []string
	if statusQuery != "" {
		for _, v := range strings.Split(statusQuery, ",") {
			if trimmed := strings.TrimSpace(v); trimmed != "" {
				statusList = append(statusList, trimmed)
			}
		}
	}
	var taskType *int
	if typeQuery := strings.TrimSpace(ctx.Query("type")); typeQuery != "" {
		value, err := strconv.Atoi(typeQuery)
		if err != nil {
			Err(ctx, errs.ParamError)
			return
		}
		taskType = &value
	}
	filters := sqllite.TaskFilters{
		Biz:    strings.TrimSpace(ctx.Query("biz")),
		Status: statusList,
		Type:   taskType,
	}
	tasks, err := service.Task.ListTasks(filters)
	if err != nil {
		Err(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, tasks)
}

func TaskGet(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		Err(ctx, errs.ParamError)
		return
	}
	task, err := service.Task.GetTask(id)
	if err != nil {
		if sqllite.IsRecordNotFound(err) {
			ctx.JSON(http.StatusNotFound, gin.H{"message": "task not found"})
			return
		}
		Err(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, task)
}

func TaskCreate(ctx *gin.Context) {
	var payload taskPayload
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		Err(ctx, err)
		return
	}
	task := domain.AppTask{
		Biz:           payload.Biz,
		Type:          payload.Type,
		Title:         payload.Title,
		Status:        payload.Status,
		StatusMsg:     payload.StatusMsg,
		StartTime:     payload.StartTime,
		EndTime:       payload.EndTime,
		ServerName:    payload.ServerName,
		ServerTitle:   payload.ServerTitle,
		ServerVersion: payload.ServerVersion,
		Param:         payload.Param,
		JobResult:     payload.JobResult,
		ModelConfig:   payload.ModelConfig,
		Result:        payload.Result,
		CreatedAt:     payload.CreatedAt,
		UpdatedAt:     payload.UpdatedAt,
	}
	created, err := service.Task.CreateTask(task)
	if err != nil {
		Err(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, created)
}

func TaskUpdate(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		Err(ctx, errs.ParamError)
		return
	}
	updates := map[string]any{}
	if err := ctx.ShouldBindJSON(&updates); err != nil {
		Err(ctx, err)
		return
	}
	normalized := normalizeTaskUpdates(updates)
	task, err := service.Task.UpdateTask(id, normalized)
	if err != nil {
		Err(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, task)
}

func TaskDelete(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		Err(ctx, errs.ParamError)
		return
	}
	if err := service.Task.DeleteTask(id); err != nil {
		Err(ctx, err)
		return
	}
	ctx.Status(http.StatusNoContent)
}

func normalizeTaskUpdates(updates map[string]any) map[string]any {
	normalized := map[string]any{}
	apply := func(key string, value any) {
		if value != nil {
			normalized[key] = value
		}
	}
	for _, key := range []string{
		"biz",
		"title",
		"status",
		"statusMsg",
		"serverName",
		"serverTitle",
		"serverVersion",
		"param",
		"jobResult",
		"modelConfig",
		"result",
	} {
		if value, ok := updates[key]; ok {
			apply(key, value)
		}
	}
	for _, key := range []string{"startTime", "endTime", "createdAt", "updatedAt"} {
		if value, ok := updates[key]; ok {
			switch typed := value.(type) {
			case float64:
				apply(key, int64(typed))
			case int64:
				apply(key, typed)
			case int:
				apply(key, int64(typed))
			}
		}
	}
	if value, ok := updates["type"]; ok {
		switch typed := value.(type) {
		case float64:
			apply("type", int(typed))
		case int:
			apply("type", typed)
		case int64:
			apply("type", int(typed))
		}
	}
	return normalized
}
