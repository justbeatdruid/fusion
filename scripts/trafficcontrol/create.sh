curl -i localhost:8001/api/v1/trafficcontrols -H 'content-type:application/json' -H 'X-auth-Token:8c06d2a4d87a8e4df280589889d85f67e6d755b2fb33538fd724ba6f700361a0' \
  -d'
{
  "data": {
    "name": "traffictest2",
    "type": "app",
    "Config": {"Second":3},
    "description": "a simple traffictest"
    
  }
}'
