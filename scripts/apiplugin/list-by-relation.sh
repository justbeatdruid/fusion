curl "localhost:8001/api/v1/applications/relation?resourceType=$1&resourceId=$2" -v \
  -H'userId:' -H'tenantId:'
