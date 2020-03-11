curl -i localhost:8001/api/v1/trafficcontrols -H 'content-type:application/json'  -H 'X-auth-Token:8c06d2a4d87a8e4df280589889d85f67e6d755b2fb33538fd724ba6f700361a0'\
  -d'
{
  "data": {
    "name": "trafficapp1",
    "type": "specialapp",
    "Config": {
        "special":[
              {"id":"078c2bc88a7e1ac1","minute":3 },
              {"id":"4088b18708c78d3f", "minute":5 }
                ]
      },
    "description": "a simple traffic control test"

  }
}'