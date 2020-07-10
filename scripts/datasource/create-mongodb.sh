#!/bin/bash

suffix=""
if test $1;then
  suffix="/connection"
fi


curl localhost:8001/api/v1/datasources${suffix} -H 'content-type:application/json' \
  -d'
{
  "data": {
    "name": "testmongo",
    "type": "mongo",
    "mongo": {
      "host": "10.160.32.24",
      "port": 27018,
      "database": "auditlogs"
      "username": "",
      "password": ""
    }
  }
}'
