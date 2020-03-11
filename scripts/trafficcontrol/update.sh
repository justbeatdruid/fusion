if ! test $1;then
  echo "need id"
  exit 1
fi



curl -i -XPATCH localhost:8001/api/v1/trafficcontrols/$1 -H 'content-type:application/json'  -H'X-auth-Token:3df7da2cea083eaab0cfbaaff9883a932cf7af92cea37f9b3b74ba9e5aee4fe8' \
  -d'
{
  "data": {
    "name": "trafficcontroltestupdate",
    "type": "app",
    "config": {"year":100,"second":4},
    "description": "a simple traffic control update test"
    
  }
}'
