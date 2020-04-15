curl -i -XPATCH localhost:8001/api/v1/restrictions/fe5167a8f1c6d985 -H 'content-type:application/json' \
-H 'userId:6' -H'tenantId:3000' -d'
{
  "data": {
    "name": "restrictionsupdatetest",
    "Config": {"ip":"11.11.11.220"}
    
  }
}'
