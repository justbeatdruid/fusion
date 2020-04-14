curl -i localhost:8001/api/v1/restrictions -H 'content-type:application/json' \
  -d'
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