if ! test $1;then
  echo "need id"
  exit 1
fi

curl localhost:8001/api/v1/serviceunits/$1/release -H 'content-type:application/json' -XPOST \
  -d'
{
  "published": true
}'

