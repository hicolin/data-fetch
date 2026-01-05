package plugin

import (
	"data-fetch/backend/plugins"
	"log"
)

func InitPluginManager() {
	manager := plugins.GetManager()
	// 假设插件已加载...

	// 1. 创建会话
	config := map[string]interface{}{"filePath": "path/to/your/data.xlsx"}
	session, err := manager.CreateSession("com.example.excel", config)
	if err != nil {
		log.Fatalf("创建会话失败: %v", err)
	}

	// 确保会话最终被销毁
	defer func() {
		if err = manager.DestroySession(session); err != nil {
			log.Printf("销毁会话时发生错误: %v", err)
		}
	}()

	// 2. 测试连接
	isConnected, err := manager.TestConnection(session)
	if err != nil {
		log.Fatalf("测试连接失败: %v", err)
	}
	if isConnected {
		log.Println("连接测试成功！")
	}

	// 3. 提取数据
	dataJSON, err := manager.FetchData(session)
	if err != nil {
		log.Fatalf("提取数据失败: %v", err)
	}

	log.Printf("成功提取数据: %s", dataJSON)
}
