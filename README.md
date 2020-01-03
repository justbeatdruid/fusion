### 代码说明

#### 1. 修改API模型
编辑xxx-controller/api/v1/xxx_types.go代码  

#### 2. 修改Controller控制逻辑
编辑xxx-controller/controller/xxx_controller.go#Reconcile函数  

#### 3. 新增路由
新增apiserver/resources/xxx/router.go函数（参考apiserver/resources/application/router.go）  
编辑apiserver/handler/handler.go，增加注册路由  

#### 4. 新增RestAPI处理逻辑
新增apiserver/resources/xxx/service（参考apiserver/resources/application/service）  