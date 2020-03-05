curl -XPATCH localhost:8001/api/v1/applications/a7d8e923bacf5ed4 -H 'content-type:application/json' \
  -H'X-auth-Token:3df7da2cea083eaab0cfbaaff9883a932cf7af92cea37f9b3b74ba9e5aee4fe8' -H'user:demo' \
  -d'
{
  "data,omitempty": {
    "name": "appsss",
    "group": "9966cdb0a1c18ca0"
  }
}'
