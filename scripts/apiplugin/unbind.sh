if ! test $1;then
  echo need apgpluginid
  exit 1
fi


curl -XPOST localhost:8001/api/v1/apiplugins/$1/apis -H 'content-type:application/json'  -v -H'tenantId:74a2512097335196b3040bed704c65c1' \
  -d'
{
  "data": {
    "operation": "unbind",
    "apis": [
    { "id": "04ee972789ea75f6"
   }

    ]
  }
}'
