package plugins

import (
	"context"
	"data-fetch/backend/svc/env"
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"
	"plugin"
	"sync"
)

// ConnectorApiVersion API版本常量
const ConnectorApiVersion = 1

// dllDir dll 文件目录
var dllDir = filepath.Join(env.RootDir, "backend", "plugins", "dll")

// Manager 管理所有已加载的插件
type Manager struct {
	plugins map[string]*ConnectorPlugin
	mutex   sync.RWMutex
}

var instance *Manager
var once sync.Once

// GetManager 获取插件管理器的单例
func GetManager() *Manager {
	once.Do(func() {
		instance = &Manager{
			plugins: make(map[string]*ConnectorPlugin),
		}
	})
	return instance
}

// LoadPlugins 扫描并加载插件目录下的所有插件
func (m *Manager) LoadPlugins(ctx context.Context) error {
	log.Printf("正在扫描插件目录: %s", dllDir)
	files, err := filepath.Glob(filepath.Join(dllDir, "*"))
	if err != nil {
		return err
	}

	for _, file := range files {
		ext := filepath.Ext(file)
		if ext != ".dll" && ext != ".so" && ext != ".dylib" {
			continue
		}

		log.Printf("正在加载插件: %s", file)
		p, err := plugin.Open(file)
		if err != nil {
			log.Printf("加载插件 %s 失败: %v", file, err)
			continue
		}

		// 1. 查找并调用 GetConnectorInfo 来获取插件基本信息
		getInfoSym, err := p.Lookup("GetConnectorInfo")
		if err != nil {
			log.Printf("在插件 %s 中找不到 'GetConnectorInfo' 符号: %v", file, err)
			continue
		}
		getInfoFunc, ok := getInfoSym.(func() ConnectorInfo)
		if !ok {
			log.Printf("插件 %s 的 'GetConnectorInfo' 符号类型不正确", file)
			continue
		}
		info := getInfoFunc()

		if info.ApiVersion != ConnectorApiVersion {
			log.Printf("插件 %s 的API版本不兼容，需要 %d, 但提供了 %d", file, ConnectorApiVersion, info.ApiVersion)
			continue
		}

		// 2. 动态查找所有需要的函数
		pluginInstance := &ConnectorPlugin{
			plugin:  p,
			info:    info,
			getInfo: getInfoFunc,
		}

		lookupAndAssign := func(name string, target interface{}) bool {
			sym, err := p.Lookup(name)
			if err != nil {
				log.Printf("在插件 %s 中找不到函数 '%s': %v", file, name, err)
				return false
			}
			// 使用类型断言
			switch t := target.(type) {
			case *func() []ConfigField:
				*t = sym.(func() []ConfigField)
			case *func(string) SessionHandle:
				*t = sym.(func(string) SessionHandle)
			case *func(SessionHandle):
				*t = sym.(func(SessionHandle))
			case *func(SessionHandle) bool:
				*t = sym.(func(SessionHandle) bool)
			case *func(SessionHandle) string:
				*t = sym.(func(SessionHandle) string)
			default:
				log.Printf("未知的目标类型 %T", target)
				return false
			}
			return true
		}

		ok = true
		ok = lookupAndAssign("GetConfigFields", &pluginInstance.getConfigFields)
		ok = ok && lookupAndAssign("CreateSession", &pluginInstance.createSession)
		ok = ok && lookupAndAssign("DestroySession", &pluginInstance.destroySession)
		ok = ok && lookupAndAssign("TestConnection", &pluginInstance.testConnection)
		ok = ok && lookupAndAssign("FetchData", &pluginInstance.fetchData)

		if !ok {
			log.Printf("插件 %s 加载不完整，已跳过", file)
			continue
		}

		m.mutex.Lock()
		m.plugins[info.ID] = pluginInstance
		m.mutex.Unlock()

		log.Printf("成功加载插件: %s (%s)", info.Name, info.ID)
	}

	return nil
}

// GetAllPlugins 返回所有已加载的插件信息
func (m *Manager) GetAllPlugins() []ConnectorInfo {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	infos := make([]ConnectorInfo, 0, len(m.plugins))
	for _, p := range m.plugins {
		infos = append(infos, p.info)
	}
	return infos
}

// GetPluginConfigFields 获取指定插件的配置项
func (m *Manager) GetPluginConfigFields(pluginID string) ([]ConfigField, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	p, exists := m.plugins[pluginID]
	if !exists {
		return nil, fmt.Errorf("插件 %s 不存在", pluginID)
	}
	return p.getConfigFields(), nil
}

// CreateSession 创建一个插件会话
func (m *Manager) CreateSession(pluginID string, config map[string]interface{}) (*Session, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	p, exists := m.plugins[pluginID]
	if !exists {
		return nil, fmt.Errorf("插件 %s 不存在", pluginID)
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("序列化配置失败: %v", err)
	}

	handle := p.createSession(string(configJSON))

	return &Session{
		PluginID: pluginID,
		Handle:   handle,
	}, nil
}

// DestroySession 销毁一个插件会话
func (m *Manager) DestroySession(session *Session) error {
	if session == nil {
		return fmt.Errorf("session 不能为空")
	}

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	p, exists := m.plugins[session.PluginID]
	if !exists {
		return fmt.Errorf("会话所属的插件 %s 不存在或已被卸载", session.PluginID)
	}

	// 调用对应插件的 destroySession 函数
	p.destroySession(session.Handle)

	log.Printf("会话 %d (插件: %s) 已被销毁。", session.Handle, session.PluginID)
	return nil
}

// TestConnection 测试连接
func (m *Manager) TestConnection(session *Session) (bool, error) {
	if session == nil {
		return false, fmt.Errorf("session 不能为空")
	}

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	p, exists := m.plugins[session.PluginID]
	if !exists {
		return false, fmt.Errorf("会话所属的插件 %s 不存在", session.PluginID)
	}

	return p.testConnection(session.Handle), nil
}

// FetchData 提取数据
func (m *Manager) FetchData(session *Session) (string, error) {
	if session == nil {
		return "", fmt.Errorf("session 不能为空")
	}

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	p, exists := m.plugins[session.PluginID]
	if !exists {
		return "", fmt.Errorf("会话所属的插件 %s 不存在", session.PluginID)
	}

	return p.fetchData(session.Handle), nil
}
