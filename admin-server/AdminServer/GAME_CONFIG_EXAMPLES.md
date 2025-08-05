# 游戏配置功能使用示例

## 场景一：游戏难度配置

### 管理后台配置

```json
// 全局配置 - 适用于所有版本
{
    "appId": "puzzle_game_001",
    "configKey": "default_lives",
    "configValue": 5,
    "configType": "number",
    "description": "默认生命数"
}

// 版本配置 - 1.0.0版本特定配置
{
    "appId": "puzzle_game_001",
    "configKey": "max_level",
    "configValue": 50,
    "version": "1.0.0",
    "configType": "number",
    "description": "1.0.0版本最大关卡数"
}

// 版本配置 - 1.1.0版本特定配置
{
    "appId": "puzzle_game_001",
    "configKey": "max_level",
    "configValue": 100,
    "version": "1.1.0",
    "configType": "number",
    "description": "1.1.0版本最大关卡数"
}
```

### 客户端使用

```javascript
// 初始化配置SDK
const gameConfig = new GameConfigSDK({
    baseURL: 'https://api.yourgame.com',
    appId: 'puzzle_game_001',
    version: '1.1.0'
});

// 游戏初始化时获取配置
async function initGame() {
    try {
        // 获取游戏配置
        const configs = await gameConfig.getConfig();
        
        // 版本配置优先：max_level = 100 (来自1.1.0版本配置)
        const maxLevel = configs.max_level?.value || 30;
        
        // 全局配置：default_lives = 5 (来自全局配置)
        const defaultLives = configs.default_lives?.value || 3;
        
        console.log(`当前版本最大关卡: ${maxLevel}`);
        console.log(`默认生命数: ${defaultLives}`);
        
        // 初始化游戏
        Game.init({
            maxLevel: maxLevel,
            defaultLives: defaultLives
        });
        
    } catch (error) {
        console.error('获取游戏配置失败，使用默认配置', error);
        
        // 降级方案：使用硬编码的默认配置
        Game.init({
            maxLevel: 30,
            defaultLives: 3
        });
    }
}
```

## 场景二：功能开关配置

### 管理后台配置

```json
// 广告功能开关
{
    "appId": "racing_game_002",
    "configKey": "enable_ads",
    "configValue": true,
    "configType": "boolean",
    "description": "是否启用广告"
}

// 新功能开关 - 仅在2.0.0版本启用
{
    "appId": "racing_game_002",
    "configKey": "enable_multiplayer",
    "configValue": true,
    "version": "2.0.0",
    "configType": "boolean",
    "description": "是否启用多人模式"
}

// 服务器维护开关
{
    "appId": "racing_game_002",
    "configKey": "server_maintenance",
    "configValue": false,
    "configType": "boolean",
    "description": "服务器维护状态"
}
```

### 客户端使用

```javascript
class FeatureManager {
    constructor(configSDK) {
        this.configSDK = configSDK;
        this.features = {};
    }
    
    async loadFeatures() {
        try {
            const configs = await this.configSDK.getConfig();
            
            this.features = {
                ads: configs.enable_ads?.value ?? true,
                multiplayer: configs.enable_multiplayer?.value ?? false,
                maintenance: configs.server_maintenance?.value ?? false
            };
            
            console.log('功能开关配置:', this.features);
            
        } catch (error) {
            console.error('加载功能开关失败:', error);
            
            // 默认配置
            this.features = {
                ads: true,
                multiplayer: false,
                maintenance: false
            };
        }
    }
    
    isFeatureEnabled(feature) {
        return this.features[feature] || false;
    }
    
    checkMaintenance() {
        if (this.features.maintenance) {
            // 显示维护提示
            showMaintenanceDialog();
            return true;
        }
        return false;
    }
}

// 使用示例
const featureManager = new FeatureManager(gameConfig);
await featureManager.loadFeatures();

// 检查服务器维护状态
if (featureManager.checkMaintenance()) {
    return; // 游戏进入维护模式
}

// 根据配置显示广告
if (featureManager.isFeatureEnabled('ads')) {
    AdManager.initialize();
}

// 根据配置显示多人模式按钮
if (featureManager.isFeatureEnabled('multiplayer')) {
    UI.showMultiplayerButton();
}
```

## 场景三：经济系统配置

### 管理后台配置

```json
// 商店配置
{
    "appId": "rpg_game_003",
    "configKey": "shop_config",
    "configValue": {
        "daily_discount": 0.8,
        "vip_discount": 0.7,
        "free_gems_per_day": 50
    },
    "configType": "object",
    "description": "商店配置"
}

// 奖励倍数配置
{
    "appId": "rpg_game_003",
    "configKey": "reward_multipliers",
    "configValue": [1.0, 1.5, 2.0, 3.0, 5.0],
    "configType": "array",
    "description": "奖励倍数列表"
}

// 版本特定的经济平衡
{
    "appId": "rpg_game_003",
    "configKey": "exp_multiplier",
    "configValue": 1.2,
    "version": "1.5.0",
    "configType": "number",
    "description": "1.5.0版本经验倍数"
}
```

### 客户端使用

```javascript
class EconomySystem {
    constructor(configSDK) {
        this.configSDK = configSDK;
        this.config = {};
    }
    
    async initialize() {
        try {
            const configs = await this.configSDK.getConfig();
            
            // 获取商店配置
            this.config.shop = configs.shop_config?.value || {
                daily_discount: 0.9,
                vip_discount: 0.8,
                free_gems_per_day: 30
            };
            
            // 获取奖励倍数
            this.config.rewardMultipliers = configs.reward_multipliers?.value || [1.0, 1.5, 2.0];
            
            // 获取经验倍数
            this.config.expMultiplier = configs.exp_multiplier?.value || 1.0;
            
            console.log('经济系统配置加载完成:', this.config);
            
        } catch (error) {
            console.error('经济系统配置加载失败:', error);
            this.loadDefaultConfig();
        }
    }
    
    loadDefaultConfig() {
        this.config = {
            shop: {
                daily_discount: 0.9,
                vip_discount: 0.8,
                free_gems_per_day: 30
            },
            rewardMultipliers: [1.0, 1.5, 2.0],
            expMultiplier: 1.0
        };
    }
    
    calculatePrice(basePrice, isVip = false, isDailyDeal = false) {
        let finalPrice = basePrice;
        
        if (isDailyDeal) {
            finalPrice *= this.config.shop.daily_discount;
        } else if (isVip) {
            finalPrice *= this.config.shop.vip_discount;
        }
        
        return Math.ceil(finalPrice);
    }
    
    calculateReward(baseReward, multiplierLevel = 0) {
        const multiplier = this.config.rewardMultipliers[multiplierLevel] || 1.0;
        return Math.floor(baseReward * multiplier);
    }
    
    calculateExp(baseExp) {
        return Math.floor(baseExp * this.config.expMultiplier);
    }
}

// 使用示例
const economySystem = new EconomySystem(gameConfig);
await economySystem.initialize();

// 计算商品价格
const swordPrice = economySystem.calculatePrice(100, player.isVip, true);
console.log(`剑的价格: ${swordPrice} 金币`);

// 计算奖励
const reward = economySystem.calculateReward(50, 2); // 使用3倍倍数
console.log(`奖励金币: ${reward}`);

// 计算经验
const exp = economySystem.calculateExp(100);
console.log(`获得经验: ${exp}`);
```

## 场景四：A/B测试配置

### 管理后台配置

```json
// A/B测试配置
{
    "appId": "strategy_game_004",
    "configKey": "ab_test_config",
    "configValue": {
        "tutorial_version": "B",
        "ui_theme": "dark",
        "battle_speed": 1.5,
        "test_groups": {
            "A": { "weight": 50, "tutorial_steps": 5 },
            "B": { "weight": 50, "tutorial_steps": 3 }
        }
    },
    "configType": "object",
    "description": "A/B测试配置"
}

// 灰度发布配置
{
    "appId": "strategy_game_004",
    "configKey": "feature_rollout",
    "configValue": {
        "new_battle_system": {
            "enabled": true,
            "rollout_percentage": 20
        },
        "social_features": {
            "enabled": false,
            "rollout_percentage": 0
        }
    },
    "configType": "object",
    "description": "功能灰度发布配置"
}
```

### 客户端使用

```javascript
class ABTestManager {
    constructor(configSDK, playerId) {
        this.configSDK = configSDK;
        this.playerId = playerId;
        this.testGroup = null;
        this.features = {};
    }
    
    async initialize() {
        try {
            const configs = await this.configSDK.getConfig();
            
            // 获取A/B测试配置
            const abConfig = configs.ab_test_config?.value;
            if (abConfig) {
                this.assignTestGroup(abConfig);
            }
            
            // 获取灰度发布配置
            const rolloutConfig = configs.feature_rollout?.value;
            if (rolloutConfig) {
                this.checkFeatureRollout(rolloutConfig);
            }
            
        } catch (error) {
            console.error('A/B测试配置加载失败:', error);
            this.testGroup = 'A'; // 默认分组
        }
    }
    
    assignTestGroup(config) {
        // 基于玩家ID的哈希值分配测试组
        const hash = this.hashCode(this.playerId);
        const random = Math.abs(hash) % 100;
        
        let cumulative = 0;
        for (const [group, data] of Object.entries(config.test_groups)) {
            cumulative += data.weight;
            if (random < cumulative) {
                this.testGroup = group;
                break;
            }
        }
        
        console.log(`玩家 ${this.playerId} 分配到测试组: ${this.testGroup}`);
    }
    
    checkFeatureRollout(rolloutConfig) {
        const hash = this.hashCode(this.playerId);
        
        for (const [feature, config] of Object.entries(rolloutConfig)) {
            if (!config.enabled) {
                this.features[feature] = false;
                continue;
            }
            
            const random = Math.abs(hash + this.hashCode(feature)) % 100;
            this.features[feature] = random < config.rollout_percentage;
        }
        
        console.log('功能开关状态:', this.features);
    }
    
    hashCode(str) {
        let hash = 0;
        for (let i = 0; i < str.length; i++) {
            const char = str.charCodeAt(i);
            hash = ((hash << 5) - hash) + char;
            hash = hash & hash; // 转换为32位整数
        }
        return hash;
    }
    
    getTestGroup() {
        return this.testGroup;
    }
    
    isFeatureEnabled(feature) {
        return this.features[feature] || false;
    }
}

// 使用示例
const abTestManager = new ABTestManager(gameConfig, player.id);
await abTestManager.initialize();

// 根据测试组显示不同的教程
const testGroup = abTestManager.getTestGroup();
if (testGroup === 'A') {
    Tutorial.start('detailed'); // 详细教程
} else {
    Tutorial.start('quick'); // 快速教程
}

// 根据灰度发布状态启用新功能
if (abTestManager.isFeatureEnabled('new_battle_system')) {
    BattleSystem.useNewVersion();
} else {
    BattleSystem.useOldVersion();
}
```

## 场景五：运营活动配置

### 管理后台配置

```json
// 限时活动配置
{
    "appId": "card_game_005",
    "configKey": "current_events",
    "configValue": {
        "double_exp_weekend": {
            "active": true,
            "start_time": "2023-12-01 00:00:00",
            "end_time": "2023-12-03 23:59:59",
            "exp_multiplier": 2.0
        },
        "holiday_sale": {
            "active": true,
            "start_time": "2023-12-20 00:00:00",
            "end_time": "2023-12-26 23:59:59",
            "discount": 0.5
        }
    },
    "configType": "object",
    "description": "当前运营活动配置"
}

// 每日任务配置
{
    "appId": "card_game_005",
    "configKey": "daily_missions",
    "configValue": [
        {
            "id": "login",
            "name": "每日登录",
            "reward": { "gold": 100, "exp": 50 },
            "required": 1
        },
        {
            "id": "battle",
            "name": "完成3场战斗",
            "reward": { "gold": 200, "cards": 1 },
            "required": 3
        },
        {
            "id": "upgrade",
            "name": "升级卡牌",
            "reward": { "gold": 150, "gems": 5 },
            "required": 1
        }
    ],
    "configType": "array",
    "description": "每日任务配置"
}
```

### 客户端使用

```javascript
class EventManager {
    constructor(configSDK) {
        this.configSDK = configSDK;
        this.activeEvents = {};
        this.dailyMissions = [];
    }
    
    async loadEvents() {
        try {
            const configs = await this.configSDK.getConfig();
            
            // 加载运营活动
            const eventsConfig = configs.current_events?.value;
            if (eventsConfig) {
                this.processEvents(eventsConfig);
            }
            
            // 加载每日任务
            this.dailyMissions = configs.daily_missions?.value || [];
            
        } catch (error) {
            console.error('活动配置加载失败:', error);
        }
    }
    
    processEvents(eventsConfig) {
        const now = new Date();
        
        for (const [eventId, eventData] of Object.entries(eventsConfig)) {
            if (!eventData.active) continue;
            
            const startTime = new Date(eventData.start_time);
            const endTime = new Date(eventData.end_time);
            
            if (now >= startTime && now <= endTime) {
                this.activeEvents[eventId] = eventData;
                console.log(`活动 ${eventId} 正在进行中`);
            }
        }
    }
    
    isEventActive(eventId) {
        return this.activeEvents.hasOwnProperty(eventId);
    }
    
    getEventData(eventId) {
        return this.activeEvents[eventId];
    }
    
    calculateExpReward(baseExp) {
        let finalExp = baseExp;
        
        // 检查双倍经验活动
        if (this.isEventActive('double_exp_weekend')) {
            const eventData = this.getEventData('double_exp_weekend');
            finalExp *= eventData.exp_multiplier;
        }
        
        return Math.floor(finalExp);
    }
    
    calculatePrice(basePrice) {
        let finalPrice = basePrice;
        
        // 检查节日促销活动
        if (this.isEventActive('holiday_sale')) {
            const eventData = this.getEventData('holiday_sale');
            finalPrice *= eventData.discount;
        }
        
        return Math.ceil(finalPrice);
    }
    
    getDailyMissions() {
        return this.dailyMissions;
    }
}

// 使用示例
const eventManager = new EventManager(gameConfig);
await eventManager.loadEvents();

// 计算经验奖励（考虑双倍经验活动）
const expReward = eventManager.calculateExpReward(100);
console.log(`经验奖励: ${expReward}`);

// 计算商品价格（考虑促销活动）
const itemPrice = eventManager.calculatePrice(50);
console.log(`商品价格: ${itemPrice} 金币`);

// 显示每日任务
const missions = eventManager.getDailyMissions();
missions.forEach(mission => {
    UI.addDailyMission(mission);
});

// 显示活动提示
if (eventManager.isEventActive('double_exp_weekend')) {
    UI.showEventNotification('双倍经验周末活动进行中！');
}
```

## 配置更新最佳实践

### 1. 热更新处理

```javascript
class ConfigManager {
    constructor(configSDK) {
        this.configSDK = configSDK;
        this.updateInterval = 5 * 60 * 1000; // 5分钟检查一次
        this.listeners = new Map();
    }
    
    startAutoUpdate() {
        setInterval(async () => {
            try {
                // 强制刷新配置
                const newConfigs = await this.configSDK.getConfig(null, false);
                this.notifyListeners(newConfigs);
            } catch (error) {
                console.error('配置自动更新失败:', error);
            }
        }, this.updateInterval);
    }
    
    onConfigUpdate(key, callback) {
        if (!this.listeners.has(key)) {
            this.listeners.set(key, []);
        }
        this.listeners.get(key).push(callback);
    }
    
    notifyListeners(newConfigs) {
        for (const [key, callbacks] of this.listeners) {
            if (newConfigs[key]) {
                callbacks.forEach(callback => {
                    callback(newConfigs[key].value);
                });
            }
        }
    }
}

// 使用示例
const configManager = new ConfigManager(gameConfig);

// 监听特定配置的变化
configManager.onConfigUpdate('max_level', (newValue) => {
    console.log(`最大关卡数更新为: ${newValue}`);
    Game.updateMaxLevel(newValue);
});

// 开始自动更新
configManager.startAutoUpdate();
```

### 2. 配置验证

```javascript
class ConfigValidator {
    static validate(configKey, value, type) {
        switch (type) {
            case 'number':
                return typeof value === 'number' && !isNaN(value);
            case 'string':
                return typeof value === 'string';
            case 'boolean':
                return typeof value === 'boolean';
            case 'object':
                return typeof value === 'object' && value !== null;
            case 'array':
                return Array.isArray(value);
            default:
                return true;
        }
    }
    
    static sanitize(configKey, value, type) {
        if (!this.validate(configKey, value, type)) {
            console.warn(`配置 ${configKey} 类型不匹配，尝试转换`);
            
            switch (type) {
                case 'number':
                    return Number(value) || 0;
                case 'string':
                    return String(value);
                case 'boolean':
                    return Boolean(value);
                default:
                    return value;
            }
        }
        
        return value;
    }
}

// 在获取配置时使用验证
const maxLevel = ConfigValidator.sanitize('max_level', configs.max_level?.value, 'number');
```

这些示例展示了游戏配置功能在实际开发中的各种应用场景，帮助开发者更好地理解和使用这个功能。 