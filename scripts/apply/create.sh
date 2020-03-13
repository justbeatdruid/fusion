curl localhost:8001/api/v1/applies -H 'content-type:application/json' -v \
  -H'X-auth-Token:3df7da2cea083eaab0cfbaaff9883a932cf7af92cea37f9b3b74ba9e5aee4fe8' -H'user:admin' \
  -d'
{
  "data": {
    "target": 
    {
      "type": "api",
      "id": "0679e262eee94fc7"
    },
    "source":
    {
      "type": "application",
      "id": "f021282e899e1d8a"
    },
    "action": "bind",
    "expireAt": "2020-11-10 15:04:05"
  }
}'
