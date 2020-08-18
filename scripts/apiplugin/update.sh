if ! test $1;then
  echo "need id"
  exit 1
fi

curl -XPUT localhost:8001/api/v1/apiplugins/$1 -H 'content-type:application/json' \
  -H'userId:1' -H'tenantId:74a2512097335196b3040bed704c65c1' \
  -d'
{
  "data": {
    "name": "apipluginsupdate",
    "description": "测试api插件更新",
    "config":{
       "remove":{
        "json":["codeupdate"],
        "headers":["xxxxupdate"]
        },
      "rename":{
    "headers":[
    {"key":"xxxxupdate", "value":"yyyyupdate"}
    ]
    },
    "replace":{
      "json":[
      {"key":"codeupdate", "value":"300update", "type":"string"}
        ],
    "headers":[
    {"key":"wwwwupdate", "value":"update"}
     ]
     },
       "append":{}
       }

  }
}'
