const factory = require("../src/api-factory");

exports.main = async (event, context) => {
    return await factory.getAPI("admin.create")(event, context);
} 