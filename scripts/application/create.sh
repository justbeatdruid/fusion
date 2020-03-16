curl localhost:8001/api/v1/applications -H 'content-type:application/json' \
  -H'X-auth-Token:3df7da2cea083eaab0cfbaaff9883a932cf7af92cea37f9b3b74ba9e5aee4fe8' -H'userId:6' -H'tenantId:10000' \
  -d'
{
  "data": {
    "name": "testtesttestapp",
    "group": "9f3acbd59a60c9e7"
  }
}'
