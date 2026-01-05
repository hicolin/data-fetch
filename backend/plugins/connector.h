#ifndef CONNECTOR_H
#define CONNECTOR_H

#ifdef __cplusplus
extern "C" {
#endif

// 版本号，用于向后兼容
#define CONNECTOR_API_VERSION 1

// 插件信息结构体
typedef struct {
  const char* id;          // 插件唯一ID
  const char* name;        // 插件名称
  const char* description; // 插件描述
  uint32_t api_version;    // 插件API版本
} ConnectorInfo;

// 配置项的定义，用于前端动态生成配置表单
typedef struct {
  const char* key;           // 配置项的键
  const char* label;         // 前端显示的标签
  const char* type;          // 输入框的类型
  const char* default_value; // 默认值
  const char* options;       // 如果是 select 类型，这里是选项的 JSON 字符串
} ConfigField;

// 导出的函数指针类型
typedef ConnectorInfo* (*GetConnectorInfoFunc)();
typedef ConfigField* (*GetConfigFieldsFunc)(int* count);
typedef void* (*CreateSessionFunc)(const char* config_json)
typedef void (*DestroySessionFunc)(void* session)
typedef bool (*TestConnectFunc)(void* session)
typedef const char* (*FetchDataFunc)(void* session)

// 插件必须导出一个名为 ConnectorPlugin 的结构体，包含所有函数指针
typedef struct {
  GetConnectorInfoFunc GetInfo;
  GetConfigFieldsFunc GetConfigFields;
  CreateSessionFunc CreateSession;
  DestroySessionFunc DestroySession;
  TestConnectFunc TestConnect;
  FetchDataFunc FetchData;
} ConnectorPlugin;

#ifdef __cplusplus
}
#endif

#endif // CONNECTOR_H