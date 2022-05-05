package server

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

	// PostInit 第二次初始化
	// 调用该函数时，所有 Module 都已完成第一段初始化过程
	// 方便进行跨Module调用
	PostInit()

	// Serve 向Bot注册服务函数
	// 结束后调用 Start
	Serve(server *Server)

	// Start 启用Module
	// 此处调用为
	// ``` go
	// go Start()
	// ```
	// 结束后正式开启服务
	Start(server *Server)

	// Stop 应用结束时对所有 Module 进行通知
	// 在此进行资源回收
	Stop(server *Server, wg *sync.WaitGroup)
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

// GetModule - 获取一个已注册的 Module 的 ModuleInfo
func GetModule(name string) (ModuleInfo, error) {
	modulesMu.Lock()
	defer modulesMu.Unlock()
	m, ok := modules[name]
	if !ok {
		return ModuleInfo{}, fmt.Errorf("module not registered: %s", name)
	}
	return m, nil
}
