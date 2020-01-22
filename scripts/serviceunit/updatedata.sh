if ! test $1;then
  echo "need id"
  exit 1
fi

curl -XPATCH localhost:8001/api/v1/serviceunits/$1 -H 'content-type:application/json' \
  -d'
{
  "data": {
    "id": "84e776c427e12f4b",
    "name": "newtest",
    "type": "data",
    "description": "a simple serviceunit",
    "datasources": [
      {
        "id": "68f510ce3b717e5c"
      }
    ]
  }
}'