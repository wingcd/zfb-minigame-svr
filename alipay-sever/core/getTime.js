exports.main = async (event, context) => {
    //返回结果
    let ret = {
        "code": 0,
        "msg": "success",
        "timestamp": Date.now()
    };

    return ret;
};