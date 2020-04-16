if ! test $1;then
  echo need appid
  exit 1
fi


curl -XPOST localhost:8001/api/v1/apis/applications/$1 -H 'content-type:application/json'  -v -H 'X-auth-Token:8c06d2a4d87a8e4df280589889d85f67e6d755b2fb33538fd724ba6f700361a0' \
  -d'
{
  "data": {
    "operation": "bind",
    "apis": [
    { "id": "7f806434c791b2ef"
   },
   {"id":"6ef34f3e5d599047" }

    ]
  }
}'