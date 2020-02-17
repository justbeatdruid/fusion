#!/bin/bash
if ! test $1;then
  echo need id
fi


curl -XPUT localhost:8001/api/v1/applies/$1/approval -H 'content-type:application/json'  -v \
  -d'
{
  "data": {
    "admitted": true,
    "reason": "no comment"
  }
}'
