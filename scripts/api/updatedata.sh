if ! test $1;then
  echo need id
  exit 1
fi

curl -XPATCH localhost:8001/api/v1/apis/$1 -H 'content-type:application/json' -H 'X-auth-Token:8c06d2a4d87a8e4df280589889d85f67e6d755b2fb33538fd724ba6f700361a0' \
  -d'
{
  "data": {
    "description": "update an description"
  }
}'