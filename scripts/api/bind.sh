if ! test $2;then
  echo need apiid, appid
  exit 1
fi


curl -XPOST localhost:8001/api/v1/apis/$1/applications/$2 -H 'content-type:application/json'  -v \
  -d'
{
  "data": {
    "operation": "release"
  }
}'
