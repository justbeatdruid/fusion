if ! test $1;then
  echo "need id"
  exit 1
fi

curl -XPATCH localhost:8001/api/v1/serviceunits/$1 -H 'content-type:application/json' \
  -d'
{
  "data": {
    "name": "sunyu1",
    "type": "web",
    "description": "a simple serviceunit",
    "kongService": {

        "host": "192.168.7.201",
        "port": 1111,
        "protocol" : "https"

     },
    "group": "182a3f334f835da7"
  }
}'
