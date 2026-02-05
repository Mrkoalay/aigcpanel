package domain

import "time"

type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type ModelServer struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Type      string            `json:"type"`
	Endpoint  string            `json:"endpoint"`
	Status    string            `json:"status"`
	Config    map[string]string `json:"config"`
	CreatedAt time.Time         `json:"createdAt"`
	UpdatedAt time.Time         `json:"updatedAt"`
}

type VoiceProfile struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Language    string    `json:"language"`
	Gender      string    `json:"gender"`
	Description string    `json:"description"`
	SamplePath  string    `json:"samplePath"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type VideoTemplate struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	PreviewPath string    `json:"previewPath"`
	VideoPath   string    `json:"videoPath"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type Task struct {
	ID          string            `json:"id"`
	Type        string            `json:"type"`
	Status      string            `json:"status"`
	Input       map[string]string `json:"input"`
	Output      map[string]string `json:"output"`
	ServerID    string            `json:"serverId"`
	OwnerUserID string            `json:"ownerUserId"`
	CreatedAt   time.Time         `json:"createdAt"`
	UpdatedAt   time.Time         `json:"updatedAt"`
}

type Database struct {
	Users          []User          `json:"users"`
	ModelServers   []ModelServer   `json:"modelServers"`
	VoiceProfiles  []VoiceProfile  `json:"voiceProfiles"`
	VideoTemplates []VideoTemplate `json:"videoTemplates"`
	Tasks          []Task          `json:"tasks"`
	AppTasks       []AppTask       `json:"appTasks"`
	Storages       []StorageRecord `json:"storages"`
	AppTemplates   []AppTemplate   `json:"appTemplates"`
}

type AppTask struct {
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
	ModelConfig   string `json:"modelConfig"`
	JobResult     string `json:"jobResult"`
	Result        string `json:"result"`
	CreatedAt     int64  `json:"createdAt"`
	UpdatedAt     int64  `json:"updatedAt"`
}

type StorageRecord struct {
	ID        int64  `json:"id"`
	Biz       string `json:"biz"`
	Sort      int64  `json:"sort"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	CreatedAt int64  `json:"createdAt"`
	UpdatedAt int64  `json:"updatedAt"`
}

type AppTemplate struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Video     string `json:"video"`
	Info      string `json:"info"`
	CreatedAt int64  `json:"createdAt"`
	UpdatedAt int64  `json:"updatedAt"`
}

type LocalModelConfigInfo struct {
	Type              string         `json:"type"`
	Name              string         `json:"name"`
	Version           string         `json:"version"`
	ServerRequire     string         `json:"serverRequire"`
	Title             string         `json:"title"`
	Description       string         `json:"description"`
	DeviceDescription string         `json:"deviceDescription"`
	Path              string         `json:"path"`
	PlatformName      string         `json:"platformName"`
	PlatformArch      string         `json:"platformArch"`
	Entry             string         `json:"entry"`
	Functions         []string       `json:"functions"`
	Settings          []any          `json:"settings"`
	Setting           map[string]any `json:"setting"`
	Config            map[string]any `json:"config"`
}
