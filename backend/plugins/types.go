package plugins

import "plugin"

// ConnectorInfo 插件信息结构体
type ConnectorInfo struct {
	ID          string `json:"id"`          // 插件唯一 ID
	Name        string `json:"name"`        // 插件名称
	Description string `json:"description"` // 插件描述
	ApiVersion  uint32 `json:"api_version"` // 插件 API 版本
}

// ConfigField 配置项的定义，用于前端动态生成配置表单
type ConfigField struct {
	Key          string `json:"key"`           // 配置项的键
	Label        string `json:"label"`         // 前端显示的标签
	Type         string `json:"type"`          // 输入框的类型
	DefaultValue string `json:"default_value"` // 默认值
	Options      string `json:"options"`       // 如果是 select 类型，这里是选项的 JSON 字符串
}

// SessionHandle 是一个不透明的句柄，代表一个插件的会话实例
// 它本质上是一个指向插件内部数据结构的指针
type SessionHandle uintptr

// Session 封装了一个插件会话的所有信息，包括它所属的插件ID和句柄
// 这是导出的，因为它需要在包外被使用
type Session struct {
	PluginID string        `json:"plugin_id"` // 插件唯一 ID
	Handle   SessionHandle `json:"handle"`    // 插件的会话句柄
}

// ConnectorPlugin 是一个已加载的插件实例，包含了所有可调用的函数
type ConnectorPlugin struct {
	plugin          *plugin.Plugin // 指向 Go plugin 包加载的插件对象
	info            ConnectorInfo
	getInfo         func() ConnectorInfo
	getConfigFields func() []ConfigField
	createSession   func(string) SessionHandle
	destroySession  func(SessionHandle)
	testConnection  func(SessionHandle) bool
	fetchData       func(SessionHandle) string
}
