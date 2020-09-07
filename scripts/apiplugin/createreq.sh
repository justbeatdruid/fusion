curl 10.160.32.24:30800/api/v1/apiplugins -H 'content-type:application/json' \
   -H'userId:1' -H'tenantId:74a2512097335196b3040bed704c65c1' \
  -d'
{
  "data": {
    "name": "孙宇req1",
    "type": "request-transformer",
    "consumerId": "4b2862af-8d18-459f-b548-35564672d4cf",
    "config":{
       "http_method": "GET",
       "remove":{
         "body":["code"],
         "headers":["xxxx"]
       },
       "rename":{
         "headers":[
         {"key":"xxxx", "value":"yyyy"}
         ]
       },
       "replace":{
         "body":[
           {"key":"code", "value":"300"}
         ],
         "headers":[
           {"key":"wwww", "value":"zzzzzzzz"}
         ]
       },
       "append":{}
       }
  },
  "description": "测试api分组创建"

}'

