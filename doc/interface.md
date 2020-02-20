## 业务汇聚平台接口文档
### 一）应用分组接口
#### 1. 创建应用分组
- 请求方式 POST
- 请求路径 /api/v1/applicationgroups
- 表头：content-type:application/json; X-auth-Token:xxxxxx

- 请求
```json
{
  "data": {
    "name": "testgroup",
    "description": "this is a test group"
  }
}
```
- 返回
```json
{
 "code": 0,
 "message": "",
 "data": {
  {
   "id": "f847df15f4c3582a",
   "namespace": "default",
   "name": "testgroup",
   "description": "this is a test group",
   "createdAt": "2020-02-10T14:50:18+08:00"
  }
}
```
- 字段说明  

| 字段 | 类型 | 说明 |  
| ---- | ---- | ---- |
| name | string | 分组名称 |
| description | string | 分组描述 |

#### 2. 删除应用分组
- 请求方式 DELETE
- 请求路径 /api/v1/applicationgroups/{id}
- 表头：content-type:application/json; X-auth-Token:xxxxxx

- 返回
```json
{
 "code": 0,
 "message": ""
}
```
- 字段说明  

| 字段 | 类型 | 说明 |  
| ---- | ---- | ---- |
| id | string | 分组ID |

#### 3. 查询应用分组列表
- 请求方式 GET
- 请求路径 /api/v1/applicationgroups
- 表头：content-type:application/json; X-auth-Token:xxxxxx

- 返回
```json
{
 "code": 0,
 "message": "",
 "data": [
{
   "id": "f847df15f4c3582a",
   "namespace": "default",
   "name": "testgroup",
   "description": "this is a test group",
   "createdAt": "2020-02-10T14:50:18+08:00"
 }
]
}
```

#### 4. 查询应用分组详情
- 请求方式 GET
- 请求路径 /api/v1/applicationgroups/{id}
- 表头：content-type:application/json; X-auth-Token:xxxxxx

- 返回
```json
{
 "code": 0,
 "message": "",
 "data": {
   "id": "f847df15f4c3582a",
   "namespace": "default",
   "name": "testgroup",
   "description": "this is a test group",
   "createdAt": "2020-02-10T14:50:18+08:00"
 }
}
```
- 字段说明  

| 字段 | 类型 | 说明 |  
| ---- | ---- | ---- |
| id | string | 分组ID |

### 二）服务单元分组接口
#### 1. 创建服务单元分组
- 请求方式 POST
- 请求路径 /api/v1/serviceunitgroups
- 表头：content-type:application/json; X-auth-Token:xxxxxx

- 请求
```json
{
  "data": {
    "name": "testgroup",
    "description": "this is a test group"
  }
}
```
- 返回
```json
{
 "code": 0,
 "message": "",
 "data": {
  {
   "id": "f847df15f4c3582a",
   "namespace": "default",
   "name": "testgroup",
   "description": "this is a test group",
   "createdAt": "2020-02-10T14:50:18+08:00"
  }
}
```
- 字段说明  

| 字段 | 类型 | 说明 |  
| ---- | ---- | ---- |
| name | string | 分组名称 |
| description | string | 分组描述 |

#### 2. 删除服务单元分组
- 请求方式 DELETE
- 请求路径 /api/v1/serviceunitgroups/{id}
- 表头：content-type:application/json; X-auth-Token:xxxxxx

- 返回
```json
{
 "code": 0,
 "message": ""
}
```
- 字段说明  

| 字段 | 类型 | 说明 |  
| ---- | ---- | ---- |
| id | string | 分组ID |

#### 3. 查询服务单元分组列表
- 请求方式 GET
- 请求路径 /api/v1/serviceunitgroups
- 表头：content-type:application/json; X-auth-Token:xxxxxx

- 返回
```json
{
 "code": 0,
 "message": "",
 "data": [
{
   "id": "f847df15f4c3582a",
   "namespace": "default",
   "name": "testgroup",
   "description": "this is a test group",
   "createdAt": "2020-02-10T14:50:18+08:00"
 }
]
}
```

#### 4. 查询服务单元分组详情
- 请求方式 GET
- 请求路径 /api/v1/serviceunitgroups/{id}
- 表头：content-type:application/json; X-auth-Token:xxxxxx

- 返回
```json
{
 "code": 0,
 "message": "",
 "data": {
   "id": "f847df15f4c3582a",
   "namespace": "default",
   "name": "testgroup",
   "description": "this is a test group",
   "createdAt": "2020-02-10T14:50:18+08:00"
 }
}
```
- 字段说明  

| 字段 | 类型 | 说明 |  
| ---- | ---- | ---- |
| id | string | 分组ID |

### 三）应用管理接口

### 四）服务单元管理接口

### 五）API管理接口

### 六）数据源管理接口

### 七）申请审批接口

