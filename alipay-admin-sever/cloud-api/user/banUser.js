const factory = require("./alipay-admin-sever/api-factory");

exports.main = async (event, context) => {
    return await factory.getAPI("user.ban")(event, context);
} 