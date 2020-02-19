curl -i localhost:8001/api/v1/trafficcontrols -H 'content-type:application/json' \
  -d'
{
  "data": {
    "name": "traffictest2",
    "type": "app",
    "Config": {"Second":3},
    "description": "a simple traffictest"
    
  }
}'
