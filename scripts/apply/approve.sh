#!/bin/bash
if ! test $1;then
  echo need id
fi


curl -XPUT localhost:8001/api/v1/applies/$1/approval -H 'content-type:application/json'  -v \
  -H'X-auth-Token:3df7da2cea083eaab0cfbaaff9883a932cf7af92cea37f9b3b74ba9e5aee4fe8' -H'user:admin' \
  -d'
{
  "data": {
    "admitted": true,
    "reason": "no comment"
  }
}'
