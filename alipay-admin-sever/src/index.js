// 导出所有函数，用于在alipay-sever/index.js中引入，接口中只需要引入层后，通过路径调用对应函数即可
require('./api-factory');

// === 管理员管理接口 ===
require('./admin/getAdminList');
require('./admin/createAdmin');
require('./admin/updateAdmin');
require('./admin/deleteAdmin');
require('./admin/resetPassword');

// === 角色管理接口 ===
require('./admin/getRoleList');
require('./admin/getAllRoles');

// === 认证相关接口 ===
require('./admin/adminLogin');
require('./admin/verifyToken');
require('./admin/initAdmin');

// === 应用管理接口 ===
require('./app/appInit');
require('./app/queryApp');
require('./app/getAllApps');
require('./app/updateApp');
require('./app/deleteApp');

// === 用户管理接口 ===
require('./user/getAllUsers');
require('./user/banUser');
require('./user/unbanUser');
require('./user/deleteUser');

// === 排行榜管理接口 ===
require('./leaderboard/getAllLeaderboards');
require('./leaderboard/updateLeaderboard');
require('./leaderboard/deleteLeaderboard');

// === 统计接口 ===
require('./stats/getDashboardStats');

// === 公共工具 ===
require('./common/auth');
require('./common/hash');