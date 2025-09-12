// 现代化 API 测试工具 JavaScript
class ApiTester {
    constructor() {
        this.config = {
            baseUrl: 'http://localhost:8081',
            appId: 'test_app',
            appSecret: 'test_secret',
            playerId: 'test_player_001'
        };

        this.stats = {
            success: 0,
            fail: 0,
            total: 0
        };

        this.testHistory = [];
        this.currentSection = 'dashboard';
        this.isRunningTests = false;

        this.init();
    }

    init() {
        this.bindEvents();
        this.loadConfig();
        this.checkConnection();
        this.initializeDefaults();
        this.updateStats();
        
        // 显示欢迎通知
        this.showNotification('Game Service API Tester 已启动', '请确保 Game Service 运行在配置的地址上', 'info');
    }

    bindEvents() {
        // 导航事件
        document.querySelectorAll('.nav-item').forEach(item => {
            item.addEventListener('click', (e) => {
                e.preventDefault();
                const section = item.dataset.section;
                this.switchSection(section);
            });
        });

        // 表单输入事件 - 实时保存配置
        ['baseUrl', 'appId', 'appSecret', 'playerId'].forEach(id => {
            const element = document.getElementById(id);
            if (element) {
                element.addEventListener('input', this.debounce(() => {
                    this.updateConfig();
                }, 500));
            }
        });

        // 键盘快捷键
        document.addEventListener('keydown', (e) => {
            if (e.ctrlKey || e.metaKey) {
                switch (e.key) {
                    case 'Enter':
                        e.preventDefault();
                        if (!this.isRunningTests) {
                            this.runAllTests();
                        }
                        break;
                    case '1':
                        e.preventDefault();
                        this.switchSection('dashboard');
                        break;
                    case '2':
                        e.preventDefault();
                        this.switchSection('config');
                        break;
                    case 'r':
                        e.preventDefault();
                        this.clearAllResults();
                        break;
                }
            }
        });
    }

    // 工具函数
    debounce(func, wait) {
        let timeout;
        return function executedFunction(...args) {
            const later = () => {
                clearTimeout(timeout);
                func(...args);
            };
            clearTimeout(timeout);
            timeout = setTimeout(later, wait);
        };
    }

    formatDuration(ms) {
        if (ms < 1000) return `${ms}ms`;
        return `${(ms / 1000).toFixed(2)}s`;
    }

    // 配置管理
    loadConfig() {
        const savedConfig = localStorage.getItem('apiTesterConfig');
        if (savedConfig) {
            try {
                this.config = { ...this.config, ...JSON.parse(savedConfig) };
                this.applyConfigToForm();
            } catch (e) {
                console.warn('加载配置失败，使用默认配置');
            }
        }
    }

    saveConfig() {
        localStorage.setItem('apiTesterConfig', JSON.stringify(this.config));
        this.showNotification('配置已保存', '配置已保存到本地存储', 'success');
    }

    updateConfig() {
        this.config.baseUrl = document.getElementById('baseUrl').value;
        this.config.appId = document.getElementById('appId').value;
        this.config.appSecret = document.getElementById('appSecret').value;
        this.config.code = document.getElementById('code').value;
        
        this.saveConfig();
        this.checkConnection();
    }

    applyConfigToForm() {
        document.getElementById('baseUrl').value = this.config.baseUrl;
        document.getElementById('appId').value = this.config.appId;
        document.getElementById('appSecret').value = this.config.appSecret;
        document.getElementById('code').value = this.config.code;
    }

    resetConfig() {
        this.config = {
            baseUrl: 'http://localhost:8081',
            appId: 'test_app',
            appSecret: 'test_secret',
            code: 'test_code'
        };
        this.applyConfigToForm();
        this.saveConfig();
        this.showNotification('配置已重置', '已恢复为默认配置', 'info');
    }

    applyPreset(type) {
        const presets = {
            local: {
                baseUrl: 'http://localhost:8081',
                appId: 'test_app',
                appSecret: 'test_secret',
                code: 'test_code'
            },
            staging: {
                baseUrl: 'https://staging-api.example.com',
                appId: 'staging_app',
                appSecret: 'staging_secret',
                code: 'staging_code'
            },
            production: {
                baseUrl: 'https://api.example.com',
                appId: 'prod_app',
                appSecret: 'prod_secret',
                code: 'prod_code'
            }
        };

        if (presets[type]) {
            this.config = { ...this.config, ...presets[type] };
            this.applyConfigToForm();
            this.saveConfig();
            this.showNotification('预设配置已应用', `已切换到${type}环境配置`, 'success');
        }
    }

    // 导航和界面
    switchSection(sectionId) {
        // 更新导航状态
        document.querySelectorAll('.nav-item').forEach(item => {
            item.classList.remove('active');
        });
        document.querySelector(`[data-section="${sectionId}"]`)?.classList.add('active');

        // 切换内容区域
        document.querySelectorAll('.content-section').forEach(section => {
            section.classList.remove('active');
        });
        document.getElementById(sectionId)?.classList.add('active');

        // 更新页面标题
        const titles = {
            dashboard: '控制台',
            config: '配置设置',
            health: '健康检查',
            user: '用户数据',
            leaderboard: '排行榜',
            counter: '计数器',
            mail: '邮件系统',
            'config-api': '配置接口'
        };
        
        document.getElementById('page-title').textContent = titles[sectionId] || '未知';
        this.currentSection = sectionId;
    }

    toggleSidebar() {
        const sidebar = document.querySelector('.sidebar');
        const mainContent = document.querySelector('.main-content');
        
        sidebar.classList.toggle('collapsed');
        mainContent.classList.toggle('expanded');
    }

    // 连接检查
    async checkConnection() {
        const statusElement = document.getElementById('connection-status');
        
        // 设置检查状态
        statusElement.className = 'connection-status checking';
        statusElement.innerHTML = '<i class="fas fa-circle animate-pulse"></i><span>检查连接中...</span>';

        try {
            const startTime = Date.now();
            const response = await fetch(`${this.config.baseUrl}/health`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' }
            });
            
            const duration = Date.now() - startTime;
            
            if (response.ok) {
                statusElement.className = 'connection-status connected';
                statusElement.innerHTML = `<i class="fas fa-circle"></i><span>已连接 (${duration}ms)</span>`;
            } else {
                throw new Error(`HTTP ${response.status}`);
            }
        } catch (error) {
            statusElement.className = 'connection-status disconnected';
            statusElement.innerHTML = '<i class="fas fa-circle"></i><span>连接失败</span>';
            console.error('连接检查失败:', error);
        }
    }

    // 通知系统
    showNotification(title, message, type = 'info', duration = 5000) {
        const container = document.getElementById('notification-container');
        const notification = document.createElement('div');
        
        const icons = {
            success: 'check-circle',
            error: 'exclamation-circle',
            warning: 'exclamation-triangle',
            info: 'info-circle'
        };

        notification.className = `notification ${type}`;
        notification.innerHTML = `
            <i class="fas fa-${icons[type]}"></i>
            <div class="notification-content">
                <div class="notification-title">${title}</div>
                <div class="notification-message">${message}</div>
            </div>
        `;

        container.appendChild(notification);

        // 显示动画
        setTimeout(() => notification.classList.add('show'), 100);

        // 自动移除
        setTimeout(() => {
            notification.classList.remove('show');
            setTimeout(() => {
                if (notification.parentNode) {
                    notification.parentNode.removeChild(notification);
                }
            }, 300);
        }, duration);
    }

    // 统计和历史
    updateStats() {
        const successRate = this.stats.total > 0 ? 
            Math.round((this.stats.success / this.stats.total) * 100) : 100;

        // 更新主界面统计
        document.getElementById('dashboard-success').textContent = this.stats.success;
        document.getElementById('dashboard-fail').textContent = this.stats.fail;
        document.getElementById('dashboard-total').textContent = this.stats.total;
        document.getElementById('dashboard-rate').textContent = `${successRate}%`;

        // 更新侧边栏统计
        document.getElementById('sidebar-success').textContent = this.stats.success;
        document.getElementById('sidebar-fail').textContent = this.stats.fail;
    }

    addToHistory(testName, success, duration, error = null) {
        const historyItem = {
            name: testName,
            success,
            duration,
            error,
            timestamp: Date.now()
        };

        this.testHistory.unshift(historyItem);
        
        // 限制历史记录数量
        if (this.testHistory.length > 50) {
            this.testHistory.pop();
        }

        this.updateHistoryDisplay();
        this.updateStats();
    }

    updateHistoryDisplay() {
        const container = document.getElementById('test-history');
        if (!container) return;

        container.innerHTML = '';

        if (this.testHistory.length === 0) {
            container.innerHTML = '<p style="color: var(--gray-500); text-align: center; padding: 20px;">暂无测试记录</p>';
            return;
        }

        this.testHistory.slice(0, 10).forEach(item => {
            const historyElement = document.createElement('div');
            historyElement.className = 'history-item';
            
            const timeAgo = this.formatTimeAgo(item.timestamp);
            
            historyElement.innerHTML = `
                <div class="history-status ${item.success ? 'success' : 'error'}"></div>
                <div class="history-info">
                    <div class="history-name">${item.name}</div>
                    <div class="history-time">${timeAgo}</div>
                </div>
                <div class="history-duration">${this.formatDuration(item.duration)}</div>
            `;

            container.appendChild(historyElement);
        });
    }

    formatTimeAgo(timestamp) {
        const diff = Date.now() - timestamp;
        const seconds = Math.floor(diff / 1000);
        const minutes = Math.floor(seconds / 60);
        const hours = Math.floor(minutes / 60);

        if (hours > 0) return `${hours}小时前`;
        if (minutes > 0) return `${minutes}分钟前`;
        return `${seconds}秒前`;
    }

    // API 请求
    async makeApiRequest(url, method = 'POST', data = {}, useAuth = true) {
        const startTime = Date.now();
        
        try {
            const timestamp = Math.floor(Date.now() / 1000);
            
            const options = {
                method: method,
                headers: {
                    'Content-Type': 'application/json',
                }
            };

            data = data || {};

            data.appId = this.config.appId;
            data.playerId = this.config.playerId;
            data.timestamp = timestamp;
            data.token = this.config.token;

            data.ver = '1.0.0';
            data.sign = generateZySign(data);
            
            if (method === 'POST' && Object.keys(data).length > 0) {                
                options.body = JSON.stringify(data);
            }

            if(url.startsWith('/')) {
                url = url.slice(1);
            }

            const fullUrl = method === 'GET' && Object.keys(data).length > 0 
                ? `${this.config.baseUrl}/${url}?${new URLSearchParams(data).toString()}`
                : `${this.config.baseUrl}/${url}`;

            const response = await fetch(fullUrl, options);
            const result = await response.json();
            const duration = Date.now() - startTime;

            return {
                success: response.ok && result.code == 0,
                data: result,
                status: response.status,
                duration
            };
        } catch (error) {
            const duration = Date.now() - startTime;
            return {
                success: false,
                error: error.message,
                data: { error: error.message },
                duration
            };
        }
    }

    // 结果显示
    displayResult(elementId, result, testName) {
        const element = document.getElementById(elementId);
        if (!element) return;

        const container = element.querySelector('.result-container') || element;
        
        // 更新统计
        this.stats.total++;
        if (result.success) {
            this.stats.success++;
        } else {
            this.stats.fail++;
        }

        // 添加到历史
        this.addToHistory(testName, result.success, result.duration, result.error);

        // 创建结果显示
        container.innerHTML = `
            <div class="result-header ${result.success ? 'success' : 'error'}">
                <span>
                    <i class="fas fa-${result.success ? 'check-circle' : 'times-circle'}"></i>
                    ${result.success ? '测试成功' : '测试失败'}
                </span>
                <span>${this.formatDuration(result.duration)}</span>
            </div>
            <div class="result-body">
                <pre>${JSON.stringify(result.data, null, 2)}</pre>
            </div>
        `;

        container.classList.add('show');

        // 显示通知
        const message = result.success ? 
            `耗时: ${this.formatDuration(result.duration)}` : 
            `错误: ${result.error || '请求失败'}`;
        
        this.showNotification(
            `${testName} - ${result.success ? '成功' : '失败'}`,
            message,
            result.success ? 'success' : 'error',
            3000
        );
    }

    // 表单填充助手
    fillSampleGameData() {
        const sampleData = {
            level: Math.floor(Math.random() * 50) + 1,
            score: Math.floor(Math.random() * 10000),
            coins: Math.floor(Math.random() * 1000),
            items: ['sword', 'shield', 'potion'],
            achievements: ['first_login', 'level_10', 'high_score'],
            lastPlayTime: new Date().toISOString(),
            settings: {
                soundEnabled: true,
                difficulty: 'normal'
            }
        };

        document.getElementById('saveDataContent').value = JSON.stringify(sampleData, null, 2);
        this.showNotification('示例数据已填充', '已生成随机游戏数据', 'info');
    }

    fillSampleUserInfo() {
        const names = ['测试玩家', '小明', '小红', '游戏高手', 'Player123'];
        const sampleUserInfo = {
            nickName: names[Math.floor(Math.random() * names.length)],
            avatarUrl: 'https://avatars.githubusercontent.com/u/1?v=4',
            gender: Math.random() > 0.5 ? 1 : 2,
            province: '测试省',
            city: '测试市',
            level: Math.floor(Math.random() * 20) + 1,
            exp: Math.floor(Math.random() * 5000),
            vipLevel: Math.floor(Math.random() * 5)
        };

        document.getElementById('userInfo').value = JSON.stringify(sampleUserInfo, null, 2);
        this.showNotification('示例信息已填充', '已生成随机用户信息', 'info');
    }

    initializeDefaults() {
        // 初始化默认值
        this.fillSampleGameData();
        this.fillSampleUserInfo();
    }

    // 显示加载状态
    showLoading(show = true) {
        const overlay = document.getElementById('loading-overlay');
        if (show) {
            overlay.classList.add('show');
        } else {
            overlay.classList.remove('show');
        }
    }

    // 清空结果
    clearAllResults() {
        document.querySelectorAll('.result-container').forEach(container => {
            container.classList.remove('show');
            setTimeout(() => {
                container.innerHTML = '';
            }, 300);
        });

        this.stats = { success: 0, fail: 0, total: 0 };
        this.testHistory = [];
        this.updateStats();
        this.updateHistoryDisplay();
        
        this.showNotification('结果已清空', '所有测试结果和历史记录已清除', 'info');
    }

    // 快速测试
    async quickTest(type) {
        const tests = {
            health: () => this.testHealth(),
            login: () => this.testLogin(),
            leaderboard: () => this.testCommitScore(),
            mail: () => this.testGetUserMails()
        };

        if (tests[type]) {
            await tests[type]();
        }
    }

    // 具体测试方法
    async testHealth() {
        const result = await this.makeApiRequest('/health', 'POST', {}, false);
        this.displayResult('health-result', result, '健康检查');
    }

    async testHeartbeat() {
        const result = await this.makeApiRequest('/heartbeat', 'POST', { playerId: this.config.playerId }, true);
        this.displayResult('heartbeat-result', result, '心跳检测');
    }

    async testLogin() {
        const code = document.getElementById('loginCode').value || 'test_code_123';
        const data = { code: code };
        
        const result = await this.makeApiRequest('/user/login', 'POST', data, true);

        // 全局保存token和playerId
        let resp = result.data.data;
        if(resp) {
            this.config.token = resp.token;
            this.config.playerId = resp.playerId;
        }

        this.displayResult('login-result', result, '用户登录');
    }

    async testSaveData() {
        const dataContent = document.getElementById('saveDataContent').value;
        let gameData;
        
        try {
            gameData = dataContent ? JSON.parse(dataContent) : { level: 1, score: 0 };
        } catch (e) {
            this.showNotification('数据格式错误', 'JSON 格式不正确，请检查数据格式', 'error');
            return;
        }
        
        const data = { data: JSON.stringify(gameData) };
        const result = await this.makeApiRequest('/user/saveData', 'POST', data, true);
        this.displayResult('saveData-result', result, '保存游戏数据');
    }

    async testGetData() {
        const result = await this.makeApiRequest('/user/getData', 'POST', {}, true);
        this.displayResult('getData-result', result, '获取游戏数据');
    }

    async testSaveUserInfo() {
        const userInfoContent = document.getElementById('userInfo').value;
        let userInfo;
        
        try {
            userInfo = userInfoContent ? JSON.parse(userInfoContent) : {
                nickName: '测试玩家',
                avatarUrl: 'http://example.com/avatar.jpg',
                gender: 1,
                level: 5
            };
        } catch (e) {
            this.showNotification('信息格式错误', 'JSON 格式不正确，请检查信息格式', 'error');
            return;
        }
        
        const data = { userInfo: JSON.stringify(userInfo) };
        const result = await this.makeApiRequest('/user/saveUserInfo', 'POST', data, true);
        this.displayResult('saveUserInfo-result', result, '保存用户信息');
    }

    async testCommitScore() {
        const type = document.getElementById('leaderboardType').value;
        const score = parseInt(document.getElementById('score').value) || 1000;
        
        const data = { type: type, score: score };
        const result = await this.makeApiRequest('/leaderboard/commit', 'POST', data, true);
        this.displayResult('commitScore-result', result, '提交分数');
    }

    async testQueryTopRank() {
        const type = document.getElementById('queryType').value;
        const count = parseInt(document.getElementById('queryCount').value) || 10;
        const startRank = parseInt(document.getElementById('startRank').value) || 1;
        
        const data = { type: type, count: count, startRank: startRank };
        const result = await this.makeApiRequest('/leaderboard/queryTopRank', 'POST', data, true);
        this.displayResult('queryTopRank-result', result, '查询排行榜');
    }

    async testIncrementCounter() {
        const counterKey = document.getElementById('counterName').value || 'test_counter';
        const location = document.getElementById('counterLocation').value || 'default';
        const increment = parseInt(document.getElementById('increment').value) || 1;
        
        const data = { 
            key: counterKey, 
            location: location,
            increment: increment 
        };
        const result = await this.makeApiRequest('/counter/increment', 'POST', data, true);
        this.displayResult('incrementCounter-result', result, '增加计数器');
    }

    async testGetCounter() {
        const counterKey = document.getElementById('getCounterName').value || 'test_counter';
        
        const data = { 
            key: counterKey,
        };
        const result = await this.makeApiRequest('/counter/get', 'POST', data, true);
        this.displayResult('getCounter-result', result, '获取计数器');
    }

    async testGetUserMails() {
        const page = parseInt(document.getElementById('mailPage').value) || 1;
        const pageSize = parseInt(document.getElementById('mailPageSize').value) || 10;
        
        const data = { page: page, pageSize: pageSize };
        const result = await this.makeApiRequest('/mail/getUserMails', 'POST', data, true);
        this.displayResult('getUserMails-result', result, '获取邮件列表');
    }

    async testUpdateMailStatus() {
        const mailId = document.getElementById('mailId').value || 'test_mail_001';
        const action = document.getElementById('mailAction').value;
        
        const data = { mailId: parseInt(mailId), status: action };
        const result = await this.makeApiRequest('/mail/updateStatus', 'POST', data, true);
        this.displayResult('updateMailStatus-result', result, '更新邮件状态');
    }

    async testGetConfig() {
        const configKey = document.getElementById('configKey').value || 'test_config';
        const data = { configKey: configKey };
        
        const result = await this.makeApiRequest('/getConfig', 'POST', data, true);
        this.displayResult('getConfig-result', result, '获取配置');
    }

    async testGetAllConfigs() {
        const result = await this.makeApiRequest('/getAllConfigs', 'POST', {}, true);
        this.displayResult('getAllConfigs-result', result, '获取所有配置');
    }

    // 批量测试
    async runAllTests() {
        if (this.isRunningTests) {
            this.showNotification('测试进行中', '请等待当前测试完成', 'warning');
            return;
        }

        this.isRunningTests = true;
        this.showLoading(true);
        
        // 重置统计
        this.stats = { success: 0, fail: 0, total: 0 };
        
        const tests = [
            { name: '用户登录', func: () => this.testLogin() },
            { name: '健康检查', func: () => this.testHealth() },
            { name: '心跳检测', func: () => this.testHeartbeat() },
            { name: '保存游戏数据', func: () => this.testSaveData() },
            { name: '获取游戏数据', func: () => this.testGetData() },
            { name: '保存用户信息', func: () => this.testSaveUserInfo() },
            { name: '提交分数', func: () => this.testCommitScore() },
            { name: '查询排行榜', func: () => this.testQueryTopRank() },
            { name: '增加计数器', func: () => this.testIncrementCounter() },
            { name: '获取计数器', func: () => this.testGetCounter() },
            { name: '获取邮件列表', func: () => this.testGetUserMails() },
            { name: '更新邮件状态', func: () => this.testUpdateMailStatus() },
            { name: '获取配置', func: () => this.testGetConfig() },
            { name: '获取所有配置', func: () => this.testGetAllConfigs() }
        ];
        
        this.showNotification('开始批量测试', `将执行 ${tests.length} 个测试用例`, 'info');
        
        for (const test of tests) {
            try {
                await test.func();
                await new Promise(resolve => setTimeout(resolve, 800)); // 延迟800ms
            } catch (error) {
                console.error(`测试 ${test.name} 失败:`, error);
            }
        }
        
        this.showLoading(false);
        this.isRunningTests = false;
        
        const successRate = Math.round((this.stats.success / this.stats.total) * 100);
        this.showNotification(
            '批量测试完成', 
            `完成 ${this.stats.total} 个测试，成功率 ${successRate}%`, 
            successRate >= 80 ? 'success' : 'warning',
            8000
        );
    }
}

// 全局函数 - 保持向后兼容
let apiTester;

// 页面加载完成后初始化
document.addEventListener('DOMContentLoaded', function() {
    apiTester = new ApiTester();
    
    // 导出全局函数
    window.toggleSidebar = () => apiTester.toggleSidebar();
    window.saveConfig = () => apiTester.saveConfig();
    window.resetConfig = () => apiTester.resetConfig();
    window.applyPreset = (type) => apiTester.applyPreset(type);
    window.fillSampleGameData = () => apiTester.fillSampleGameData();
    window.fillSampleUserInfo = () => apiTester.fillSampleUserInfo();
    window.quickTest = (type) => apiTester.quickTest(type);
    window.clearAllResults = () => apiTester.clearAllResults();
    window.runAllTests = () => apiTester.runAllTests();
    
    // 测试函数
    window.testHealth = () => apiTester.testHealth();
    window.testHeartbeat = () => apiTester.testHeartbeat();
    window.testLogin = () => apiTester.testLogin();
    window.testSaveData = () => apiTester.testSaveData();
    window.testGetData = () => apiTester.testGetData();
    window.testSaveUserInfo = () => apiTester.testSaveUserInfo();
    window.testCommitScore = () => apiTester.testCommitScore();
    window.testQueryTopRank = () => apiTester.testQueryTopRank();
    window.testIncrementCounter = () => apiTester.testIncrementCounter();
    window.testGetCounter = () => apiTester.testGetCounter();
    window.testGetUserMails = () => apiTester.testGetUserMails();
    window.testUpdateMailStatus = () => apiTester.testUpdateMailStatus();
    window.testGetConfig = () => apiTester.testGetConfig();
    window.testGetAllConfigs = () => apiTester.testGetAllConfigs();
});