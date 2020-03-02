if ! test $1;then
  echo need id
  exit 1
fi

curl -XPOST localhost:8001/api/v1/serviceunits/$1/users -v \
  -H 'content-type:application/json' -H'X-auth-Token:3df7da2cea083eaab0cfbaaff9883a932cf7af92cea37f9b3b74ba9e5aee4fe8' -H'user:admin' \
  -d'{"data":{"id": "testuser", "role": "manager"}}'
