#!/bin/bash

suffix=""
if test $1;then
  suffix="/connection"
fi


curl localhost:8001/api/v1/datasources${suffix} -H 'content-type:application/json' \
  -d'
{
  "data": {
    "name": "testhive",
    "type": "hive",
    "hive": {
      "host": "10.160.32.24",
      "port": 10000,
      "database": "test",
      "hdfsPath": "hdfs://node2:10000/xxx/xxx/xxx.db",
      "defaultFs": "hdfs://node2:9000",
      "jdbcUrl": "jdbc:hive2://node2:10000/hdfswriter"
    }
  }
}'
