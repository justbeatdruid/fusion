curl localhost:8001/api/v1/serviceunits -H 'content-type:application/json' \
  -H'X-auth-Token:3df7da2cea083eaab0cfbaaff9883a932cf7af92cea37f9b3b74ba9e5aee4fe8' -H'user:demo' \
  -d'
{
  "data": {
    "name": "newnewnewtest",
    "type": "data",
    "description": "a simple serviceunit",
    "datasources":
      {
	"id": "c13e8fb8ab809ddf"
      },
    "group": "182a3f334f835da7"
  }
}'
