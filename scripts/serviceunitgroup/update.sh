if ! test $1;then
  echo need id
  exit 1
fi

curl localhost:8001/api/v1/serviceunitgroups/$1 -H 'content-type:application/json' \
  -H'X-auth-Token:3df7da2cea083eaab0cfbaaff9883a932cf7af92cea37f9b3b74ba9e5aee4fe8' -H'user:admin' \
  -XPUT -d'
{
  "data": {
    "name": "财务室",
    "description": "this is a test serviceunit group"
  }
}'
