if ! test $1;then
  echo need appid
  exit 1
fi


curl -XPOST localhost:8001/api/v1/apis/applications/$1 -H 'content-type:application/json'  -v -H'tenantId:74a2512097335196b3040bed704c65c1' \
  -d'
{
  "data": {
    "operation": "bind",
    "apis": [
    { "id": "04ee972789ea75f6"
   },
   {"id":"6ef34f3e5d599047" }

    ]
  }
}'