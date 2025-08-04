// 测试邮件函数
console.log('开始测试邮件功能...');

try {
    // 测试加载sendMail函数
    const sendMail = require('./cloud/functions/sendMail');
    console.log('✓ sendMail函数加载成功');
    
    // 测试加载createMail函数
    const createMail = require('./cloud/functions/createMail');
    console.log('✓ createMail函数加载成功');
    
    // 测试加载getAllMails函数
    const getAllMails = require('./cloud/functions/getAllMails');
    console.log('✓ getAllMails函数加载成功');
    
    // 测试加载getMailStats函数
    const getMailStats = require('./cloud/functions/getMailStats');
    console.log('✓ getMailStats函数加载成功');
    
    console.log('✅ 所有邮件函数加载成功！');
    
} catch (error) {
    console.error('❌ 邮件函数加载失败:', error.message);
    console.error('错误详情:', error);
} 