curl localhost:8001/api/v1/applies -H 'content-type:application/json' -v \
  -d'
{
  "data": {
    "target": 
    {
      "type": "api",
      "id": "895c74b50e4ebd25"
    },
    "source":
    {
      "type": "application",
      "id": "190d510db16e80b7"
    },
    "action": "bind",
    "expireAt": "2020-10-10T15:04:05+08:00"
  }
}'
