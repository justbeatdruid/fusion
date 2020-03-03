curl localhost:8001/api/v1/applies -H 'content-type:application/json' -v \
  -H'X-auth-Token:3df7da2cea083eaab0cfbaaff9883a932cf7af92cea37f9b3b74ba9e5aee4fe8' -H'user:admin' \
  -d'
{
  "data": {
    "target": 
    {
      "type": "api",
      "id": "65a391d7cbedd6bf"
    },
    "source":
    {
      "type": "application",
      "id": "97e467eb2b734142"
    },
    "action": "bind",
    "expireAt": "2020-11-10T15:04:05+08:00"
  }
}'
