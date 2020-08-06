if ! test $1;then
  echo need id
fi

curl -XPOST localhost:8001/api/v1/apis/$1/plugins -H 'content-type:application/json'   -H'tenantId:74a2512097335196b3040bed704c65c1'  -v \
 -d'
{
  "data": {
       "name":"response-transformer",
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

  }
}'
