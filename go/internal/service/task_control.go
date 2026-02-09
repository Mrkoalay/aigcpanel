package service

import (
	"fmt"
	"sync"
	"xiacutai-server/internal/component/modelcall/easyserver"
)

var taskServerRegistry = struct {
	sync.Mutex
	servers map[int64]*easyserver.EasyServer
}{
	servers: make(map[int64]*easyserver.EasyServer),
}

func registerTaskServer(taskID int64, server *easyserver.EasyServer) {
	if server == nil {
		return
	}
	taskServerRegistry.Lock()
	defer taskServerRegistry.Unlock()
	taskServerRegistry.servers[taskID] = server
}

func unregisterTaskServer(taskID int64) {
	taskServerRegistry.Lock()
	defer taskServerRegistry.Unlock()
	delete(taskServerRegistry.servers, taskID)
}

func CancelEasyServerTask(taskID int64) error {
	taskServerRegistry.Lock()
	server, ok := taskServerRegistry.servers[taskID]
	taskServerRegistry.Unlock()
	if !ok || server == nil {
		return fmt.Errorf("task %d is not running", taskID)
	}
	return server.Cancel()
}
