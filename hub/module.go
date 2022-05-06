package hub

import (
	"fmt"
	"sync"
)

var (
	modules   = make(map[string]ModuleInfo)
	modulesMu sync.RWMutex
)

type Module interface {
	GetModuleInfo() ModuleInfo

	// Module 的生命周期

	// Init 初始化
	// 待所有 Module 初始化完成后
	// 进行服务注册 Serve
	Init()

	// Serve 注册服务函数
	Serve(server *Server)
}

// RegisterModule - 向全局添加 Module
func RegisterModule(instance Module) {
	mod := instance.GetModuleInfo()

	if mod.ID == "" {
		panic("module ID missing")
	}
	if mod.Instance == nil {
		panic("missing ModuleInfo.Instance")
	}

	modulesMu.Lock()
	defer modulesMu.Unlock()

	if _, ok := modules[string(mod.ID)]; ok {
		panic(fmt.Sprintf("module already registered: %s", mod.ID))
	}
	modules[string(mod.ID)] = mod
}
