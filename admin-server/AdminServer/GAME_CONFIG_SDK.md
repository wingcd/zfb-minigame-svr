# 游戏配置客户端 SDK 使用指南

## 概述

游戏配置SDK提供了简单易用的接口，让游戏客户端能够方便地获取远程配置。支持版本配置优先级、本地缓存等功能。

## JavaScript SDK

### 基础用法

```javascript
class GameConfigSDK {
    constructor(options) {
        this.baseURL = options.baseURL;
        this.appId = options.appId;
        this.version = options.version || null;
        this.cache = new Map();
        this.cacheExpireTime = options.cacheExpireTime || 5 * 60 * 1000; // 5分钟
    }

    /**
     * 获取配置
     * @param {string} configKey - 配置键名（可选）
     * @param {boolean} useCache - 是否使用缓存
     * @returns {Promise<Object>} 配置对象
     */
    async getConfig(configKey = null, useCache = true) {
        const cacheKey = `${this.appId}_${this.version}_${configKey || 'all'}`;
        
        // 检查缓存
        if (useCache && this.cache.has(cacheKey)) {
            const cached = this.cache.get(cacheKey);
            if (Date.now() - cached.timestamp < this.cacheExpireTime) {
                return cached.data;
            }
        }

        try {
            const params = {
                appId: this.appId
            };
            
            if (this.version) {
                params.version = this.version;
            }
            
            if (configKey) {
                params.configKey = configKey;
            }

            const response = await fetch(`${this.baseURL}/gameConfig/get`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(params)
            });

            const result = await response.json();
            
            if (result.code === 0) {
                const configs = result.data.configs;
                
                // 缓存结果
                this.cache.set(cacheKey, {
                    data: configs,
                    timestamp: Date.now()
                });
                
                return configs;
            } else {
                throw new Error(result.msg || '获取配置失败');
            }
        } catch (error) {
            console.error('获取游戏配置失败:', error);
            throw error;
        }
    }

    /**
     * 获取单个配置值
     * @param {string} configKey - 配置键名
     * @param {any} defaultValue - 默认值
     * @returns {Promise<any>} 配置值
     */
    async getValue(configKey, defaultValue = null) {
        try {
            const configs = await this.getConfig(configKey);
            return configs[configKey]?.value ?? defaultValue;
        } catch (error) {
            console.warn(`获取配置 ${configKey} 失败，使用默认值:`, defaultValue);
            return defaultValue;
        }
    }

    /**
     * 批量获取配置值
     * @param {Array<string>} configKeys - 配置键名数组
     * @returns {Promise<Object>} 配置值对象
     */
    async getValues(configKeys) {
        try {
            const configs = await this.getConfig();
            const result = {};
            
            configKeys.forEach(key => {
                result[key] = configs[key]?.value ?? null;
            });
            
            return result;
        } catch (error) {
            console.error('批量获取配置失败:', error);
            const result = {};
            configKeys.forEach(key => {
                result[key] = null;
            });
            return result;
        }
    }

    /**
     * 清除缓存
     */
    clearCache() {
        this.cache.clear();
    }

    /**
     * 预加载配置
     */
    async preload() {
        try {
            await this.getConfig();
            console.log('配置预加载完成');
        } catch (error) {
            console.error('配置预加载失败:', error);
        }
    }
}
```

### 使用示例

```javascript
// 初始化SDK
const gameConfig = new GameConfigSDK({
    baseURL: 'https://your-api-domain.com',
    appId: 'your_game_id',
    version: '1.0.0',
    cacheExpireTime: 10 * 60 * 1000 // 10分钟缓存
});

// 游戏启动时预加载配置
await gameConfig.preload();

// 获取单个配置
const maxLevel = await gameConfig.getValue('max_level', 30);
const serverUrl = await gameConfig.getValue('server_url', 'https://default.com');

// 批量获取配置
const configs = await gameConfig.getValues([
    'max_level',
    'server_url',
    'enable_ads',
    'reward_multiplier'
]);

console.log('最大关卡:', configs.max_level);
console.log('服务器地址:', configs.server_url);
console.log('启用广告:', configs.enable_ads);
console.log('奖励倍数:', configs.reward_multiplier);

// 获取所有配置
const allConfigs = await gameConfig.getConfig();
Object.keys(allConfigs).forEach(key => {
    const config = allConfigs[key];
    console.log(`${key}: ${config.value} (来源: ${config.source})`);
});
```

## 微信小程序 SDK

```javascript
class WechatGameConfigSDK {
    constructor(options) {
        this.baseURL = options.baseURL;
        this.appId = options.appId;
        this.version = options.version || null;
        this.storageKey = 'game_config_cache';
        this.cacheExpireTime = options.cacheExpireTime || 5 * 60 * 1000;
    }

    async getConfig(configKey = null, useCache = true) {
        const cacheKey = `${this.appId}_${this.version}_${configKey || 'all'}`;
        
        // 检查缓存
        if (useCache) {
            try {
                const cached = wx.getStorageSync(this.storageKey) || {};
                if (cached[cacheKey] && Date.now() - cached[cacheKey].timestamp < this.cacheExpireTime) {
                    return cached[cacheKey].data;
                }
            } catch (error) {
                console.warn('读取缓存失败:', error);
            }
        }

        return new Promise((resolve, reject) => {
            const params = { appId: this.appId };
            
            if (this.version) params.version = this.version;
            if (configKey) params.configKey = configKey;

            wx.request({
                url: `${this.baseURL}/gameConfig/get`,
                method: 'POST',
                data: params,
                success: (res) => {
                    if (res.data.code === 0) {
                        const configs = res.data.data.configs;
                        
                        // 缓存结果
                        try {
                            const cached = wx.getStorageSync(this.storageKey) || {};
                            cached[cacheKey] = {
                                data: configs,
                                timestamp: Date.now()
                            };
                            wx.setStorageSync(this.storageKey, cached);
                        } catch (error) {
                            console.warn('缓存失败:', error);
                        }
                        
                        resolve(configs);
                    } else {
                        reject(new Error(res.data.msg || '获取配置失败'));
                    }
                },
                fail: reject
            });
        });
    }

    async getValue(configKey, defaultValue = null) {
        try {
            const configs = await this.getConfig(configKey);
            return configs[configKey]?.value ?? defaultValue;
        } catch (error) {
            console.warn(`获取配置 ${configKey} 失败，使用默认值:`, defaultValue);
            return defaultValue;
        }
    }

    clearCache() {
        try {
            wx.removeStorageSync(this.storageKey);
        } catch (error) {
            console.error('清除缓存失败:', error);
        }
    }
}
```

## Unity C# SDK

```csharp
using System;
using System.Collections.Generic;
using UnityEngine;
using UnityEngine.Networking;
using System.Collections;

public class GameConfigSDK : MonoBehaviour
{
    [System.Serializable]
    public class ConfigResponse
    {
        public int code;
        public string msg;
        public ConfigData data;
    }

    [System.Serializable]
    public class ConfigData
    {
        public Dictionary<string, ConfigItem> configs;
    }

    [System.Serializable]
    public class ConfigItem
    {
        public object value;
        public string type;
        public string source;
        public string version;
        public string description;
    }

    public string baseURL;
    public string appId;
    public string version;
    
    private Dictionary<string, ConfigItem> cachedConfigs;
    private float cacheTime;
    private float cacheExpireTime = 300f; // 5分钟

    public void Initialize(string baseURL, string appId, string version = null)
    {
        this.baseURL = baseURL;
        this.appId = appId;
        this.version = version;
    }

    public IEnumerator GetConfig(System.Action<Dictionary<string, ConfigItem>> onSuccess, System.Action<string> onError, string configKey = null)
    {
        // 检查缓存
        if (cachedConfigs != null && Time.time - cacheTime < cacheExpireTime)
        {
            onSuccess?.Invoke(cachedConfigs);
            yield break;
        }

        var requestData = new Dictionary<string, object>
        {
            ["appId"] = appId
        };

        if (!string.IsNullOrEmpty(version))
            requestData["version"] = version;

        if (!string.IsNullOrEmpty(configKey))
            requestData["configKey"] = configKey;

        string jsonData = JsonUtility.ToJson(requestData);
        
        using (UnityWebRequest request = new UnityWebRequest($"{baseURL}/gameConfig/get", "POST"))
        {
            byte[] bodyRaw = System.Text.Encoding.UTF8.GetBytes(jsonData);
            request.uploadHandler = new UploadHandlerRaw(bodyRaw);
            request.downloadHandler = new DownloadHandlerBuffer();
            request.SetRequestHeader("Content-Type", "application/json");

            yield return request.SendWebRequest();

            if (request.result == UnityWebRequest.Result.Success)
            {
                try
                {
                    ConfigResponse response = JsonUtility.FromJson<ConfigResponse>(request.downloadHandler.text);
                    
                    if (response.code == 0)
                    {
                        cachedConfigs = response.data.configs;
                        cacheTime = Time.time;
                        onSuccess?.Invoke(cachedConfigs);
                    }
                    else
                    {
                        onError?.Invoke(response.msg ?? "获取配置失败");
                    }
                }
                catch (Exception e)
                {
                    onError?.Invoke($"解析响应失败: {e.Message}");
                }
            }
            else
            {
                onError?.Invoke($"网络请求失败: {request.error}");
            }
        }
    }

    public T GetValue<T>(string configKey, T defaultValue = default(T))
    {
        if (cachedConfigs != null && cachedConfigs.ContainsKey(configKey))
        {
            try
            {
                return (T)Convert.ChangeType(cachedConfigs[configKey].value, typeof(T));
            }
            catch
            {
                Debug.LogWarning($"配置 {configKey} 类型转换失败，使用默认值");
            }
        }
        
        return defaultValue;
    }

    public void ClearCache()
    {
        cachedConfigs = null;
        cacheTime = 0;
    }
}
```

### Unity 使用示例

```csharp
public class GameManager : MonoBehaviour
{
    public GameConfigSDK configSDK;
    
    void Start()
    {
        configSDK.Initialize("https://your-api-domain.com", "your_game_id", "1.0.0");
        StartCoroutine(LoadConfigs());
    }
    
    IEnumerator LoadConfigs()
    {
        yield return StartCoroutine(configSDK.GetConfig(
            onSuccess: (configs) => {
                // 配置加载成功
                int maxLevel = configSDK.GetValue<int>("max_level", 30);
                string serverUrl = configSDK.GetValue<string>("server_url", "https://default.com");
                bool enableAds = configSDK.GetValue<bool>("enable_ads", true);
                
                Debug.Log($"最大关卡: {maxLevel}");
                Debug.Log($"服务器地址: {serverUrl}");
                Debug.Log($"启用广告: {enableAds}");
            },
            onError: (error) => {
                Debug.LogError($"加载配置失败: {error}");
            }
        ));
    }
}
```

## 最佳实践

### 1. 缓存策略
- 合理设置缓存过期时间（建议5-10分钟）
- 在游戏启动时预加载配置
- 提供离线降级方案

### 2. 错误处理
- 始终提供默认值
- 优雅处理网络错误
- 记录配置获取失败的日志

### 3. 性能优化
- 批量获取配置而非单个获取
- 避免频繁请求配置
- 使用本地存储缓存配置

### 4. 版本管理
- 明确指定游戏版本
- 合理规划配置的版本策略
- 考虑向后兼容性

### 5. 安全考虑
- 不要在配置中存储敏感信息
- 验证配置值的合法性
- 防止配置注入攻击 