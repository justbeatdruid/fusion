if ! test $1;then
  echo "need id"
  exit 1
fi

curl -XPATCH localhost:8001/api/v1/serviceunits/$1 -H 'content-type:application/json' \
  -d'
{
  "data": {
    "name": "sssssyyyy",
    "type": "data",
    "description": "simple serviceunit",
    "datasources":
      {
        "id": "caeaa2f64e39cdcf"
      },
    "group": "182a3f334f835da7"
  }
}'
