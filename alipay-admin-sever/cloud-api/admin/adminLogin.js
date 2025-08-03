const { getAPI } = require("./src/api-factory");
exports.main = async (event, context) => {
    return await getAPI("auth.login")(event, context);
} 