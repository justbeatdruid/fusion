
if ! test $1;then
  echo "need id"
  exit 1
fi



curl -i -XPATCH localhost:8001/api/v1/trafficcontrols/$1  -H 'content-type:application/json'  -H 'X-auth-Token:8c06d2a4d87a8e4df280589889d85f67e6d755b2fb33538fd724ba6f700361a0'\
-H 'userId:6' -H'tenantId:3000'  -d'
{
  "data": {
    "name": "trafficapp1",
    "type": "specialapp",
    "config": {
        "special":[
	      {"id":"4088b18708c78d3f", "minute":6 },
	      {"id":"33333", "minute":6 }	      
	        ]
      },
    "description": "a simple traffic control test"
    
  }
}'
