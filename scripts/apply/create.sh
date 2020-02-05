curl localhost:8001/api/v1/applies -H 'content-type:application/json' -v \
  -d'
{
  "data": {
    "name": "xuxutest",
    "target": 
    {
      "type": "api",
      "id": "edbea8c8055ef541",
      "name": "???"
    },
    "source":
    {
      "type": "application",
      "id": "85de4d6bf0542989",
      "name": "???"
    },
    "action": "bind",
    "expireAt": "2020-10-10T15:04:05+08:00"
  }
}'
