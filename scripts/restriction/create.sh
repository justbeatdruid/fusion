curl -i localhost:8001/api/v1/restrictions -H 'content-type:application/json' \
-H 'userId:6' -H'tenantId:3000' -d'
{
  "data": {
    "name": "restrictions",
    "type": "ip",
    "action": "white",
    "Config": {
      "ip": [
        "11.11.11.11",
        "12.12.12.12"
      ]
    }
  }
}'