curl -i localhost:8001/api/v1/trafficcontrols -H 'content-type:application/json'  -H'X-auth-Token:3df7da2cea083eaab0cfbaaff9883a932cf7af92cea37f9b3b74ba9e5aee4fe8' \
-H 'userId:6' -H'tenantId:3000' -d'
{
  "data": {
    "name": "wmjtrafficupdate",
    "type": "app",
    "config": {"year":100,"second":3},
    "description": "a simple traffic control test"

  }
}'

