curl localhost:8001/api/v1/apply/create -H 'content-type:application/json' -v \
  -d'
{
  "data": {
    "name": "xuxutest",
    "targetType": "api",
    "targetID": "51e6c509de3b1bf",
    "targetName": "???",
    "appID": "85de4d6bf0542989",
    "appName": "???",
    "expireAt": "2020-10-10T15:04:05+08:00"
  }
}'
