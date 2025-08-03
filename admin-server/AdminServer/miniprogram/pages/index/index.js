Page({
  data:{
    data:null
  },
  async onLoad(){
    //初始化 context ，这里代码可以抽象为公共代码
    const context = await my.cloud.createCloudContext({
      env: 'env-00jxt0uhcb2h' //修改为自己的环境 ID 
    });
    
    await context.init();
    var self = this;
    my.showLoading({
      content: '加载中...',
      delay: '100',
    });
    console.log(my.fncontext);
    context.callFunction({
      name:'helloworld',
      success:function(res){
         my.hideLoading();
         console.log(res);
         self.setData({
           data:res.result.message
         });
      },
      fail:function(erro){
        my.hideLoading();
        console.log(erro);
        self.setData({
          data:null
        });
      }
    });
  },
})