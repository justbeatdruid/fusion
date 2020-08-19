if ! test $1;then
  echo need apgroueid
  exit 1
fi


curl -XPOST localhost:8001/api/v1/apigroups/$1/apis -H 'content-type:application/json'  -v -H'tenantId:74a2512097335196b3040bed704c65c1' \
  -d'
{
  "data": {
    "operation": "bind",
    "apis": [
    { "id": "04ee972789ea75f6"
   },
   {
      "id":"8e8b46aba2b97be9"
    },
    {
      "id":"4601ae17af7328ca"
    }

    ]
  }
}'