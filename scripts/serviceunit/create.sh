curl localhost:8001/api/v1/serviceunits -H 'content-type:application/json' \
  -H'X-auth-Token:3df7da2cea083eaab0cfbaaff9883a932cf7af92cea37f9b3b74ba9e5aee4fe8' -H'userId:6' -H'tenantId:3000'\
  -d'
{
  "data": {
    "name": "newnewnewtest",
    "type": "data",
    "description": "a simple serviceunit",
    "datasources":
      {
	"id": "4f1d1b902c1e663a"
      },
    "group": "ae89da1879ae1b95"
  }
}'
