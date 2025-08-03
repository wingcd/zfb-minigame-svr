const apiList = {};

exports.registerAPI = (apiName, func) => {
    apiList[apiName] = func;
}

exports.getAPI = (apiName) => {
    return apiList[apiName];
}

exports.getAPIList = () => {
    return apiList;
}

exports.autoRegister = (apiName) => {
    return (targetFunc) => {
        exports.registerAPI(apiName, targetFunc);
        return targetFunc;
    }
}