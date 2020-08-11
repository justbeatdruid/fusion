if ! test $1;then
  echo "need id"
  exit 1
fi

curl -XPUT localhost:8001/api/v1/apigroups/$1 -H 'content-type:application/json' \
  -H'userId:1' -H'tenantId:74a2512097335196b3040bed704c65c1' \
  -d'
{
  "data": {
    "name": "apigroupsupdate",
    "description": "测试api分组更新"
  }
}'
