curl -i -XPOST  localhost:8001/api/v1/api/test -H 'content-type:application/json' -H 'X-auth-Token:8c06d2a4d87a8e4df280589889d85f67e6d755b2fb33538fd724ba6f700361a0' \
  -d'
  {
 "data": {
  "id": "16d1719a951985da",
  "namespace": "default",
  "name": "调试api测试",
  "description": "调试api",
  "serviceunit": {
   "id": "5cfe117f59066122",
   "name": "调试集群一",
   "group": "",
   "kongID": "6a8134a2-692a-4a91-afa2-a8d92b4ba962",
   "Type": "web",
   "Host": "10.160.32.31",
   "Port": 30800,
   "protocol": "http"
  },
  "applications": [],
  "users": {
   "owner": {
    "id": "0",
    "name": "system"
   },
   "managers": [],
   "members": []
  },
  "frequency": 10,
  "apiType": "public",
  "authType": "NOAUTR",
  "tags": "testapi",
  "apiBackendType": "web",
  "method": "GET",
  "protocol": "HTTP",
  "returnType": "json",
  "apiRequestParameters": null,
  "apiResponseParameters": null,
  "apiPublicParameters": null,
  "apiDefineInfo": {
   "path": "/api/v1/apis",
   "matchMode": "2",
   "method": "GET",
   "protocol": "HTTP",
   "cors": "false"
  },
  "kongApi": {
   "hosts": null,
   "paths": [
    "/api/v1/apis"
   ],
   "headers": null,
   "methods": [
    "GET"
   ],
   "https_redirect_status_code": 0,
   "regex_priority": 0,
   "strip_path": false,
   "preserve_host": false,
   "snis": null,
   "protocols": [
    "http"
   ],
   "kong_id": "17f4d480-6f6b-4b8c-9e76-5e5b35ca1efd",
   "jwt_id": "",
   "acl_id": "",
   "cors_id": "",
   "prometheus_id": "d718700c-050f-42d3-be75-bda6777f1cd4"
  },
  "apiQueryInfo": {
   "webParams": []
  },
  "apiReturnInfo": {
   "normalExample": "{”code“:0}",
   "failureExample": ""
  },
  "traffic": {
   "id": "",
   "name": "",
   "specialID": null
  },
  "restriction": {
   "id": "",
   "name": ""
  },
  "status": "success",
  "action": "publish",
  "publishStatus": "released",
  "access": "http://kong-kong-admin:30080/api/v1/apis",
  "updatedAt": "2020-04-24 02:06:03",
  "releasedAt": "2020-04-24 02:06:03",
  "applicationCount": 0,
  "calledCount": 0,
  "PublishInfo": {
   "version": "80188971911b940e",
   "host": "",
   "publishCount": 1
  },
  "applicationBindStatus": null
 }
}'

