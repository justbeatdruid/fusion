#!/bin/bash
set -ex
curl localhost:8001/api/v1/apiplugins -H 'content-type:application/json' \
   -H'userId:1' -H'tenantId:74a2512097335196b3040bed704c65c1' \
  -d'
{
  "data": {
    "name": "apiplugin1",
    "type": "response-transformer",
    "config":{
       "remove":{
        "json":["code"],
        "headers":["xxxx"]
        },
      "rename":{
    "headers":[
    {"key":"xxxx", "value":"yyyy"}
    ]
    },
    "replace":{
      "json":[
      {"key":"code", "value":"300", "type":"string"}
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

