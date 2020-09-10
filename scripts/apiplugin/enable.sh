if ! test $1;then
  echo need apgpluginid
  exit 1
fi



curl -XPOST localhost:8001/api/v1/apiplugins/$1/relations -H 'content-type:application/json'  -v -H'tenantId:74a2512097335196b3040bed704c65c1' \
  -d'
{
  "data": {
    "enable": true,
    "ids": [
    { "id": "ea498fa6-7179-42e3-b8e7-d1bd47b569fe"
   }
    ]
  }
}'