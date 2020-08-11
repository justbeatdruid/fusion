#!/bin/bash
set -ex
curl localhost:8001/api/v1/apigroups -H 'content-type:application/json' \
   -H'userId:1' -H'tenantId:74a2512097335196b3040bed704c65c1' \
  -d'
{
  "data": {
    "name": "apigroups1",
    "description": "测试api分组创建"
  }
}'

