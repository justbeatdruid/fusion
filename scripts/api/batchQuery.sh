
curl -XPOST localhost:8001/api/v1/apis/capabilities -H 'content-type:application/json'  -v  -H 'userId:1'  \
  -d'
{
  "data": {
    "apis": [ "04ee972789ea75f6" , "8e8b46aba2b97be9", "4601ae17af7328ca"]
  }
}'