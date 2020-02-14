curl localhost:8001/api/v1/applies -H 'content-type:application/json' -v \
  -d'
{
  "data": {
    "target": 
    {
      "type": "api",
      "id": "003d2691017b1056",
      "name": "???"
    },
    "source":
    {
      "type": "application",
      "id": "078c2bc88a7e1ac1",
      "name": "???"
    },
    "action": "bind",
    "expireAt": "2020-10-10T15:04:05+08:00"
  }
}'
