const admin = require("./alipay-admin-sever");

exports.main = async (event, context) => {
    return await admin.callAdminAPI("user.delete", event, context);
} 