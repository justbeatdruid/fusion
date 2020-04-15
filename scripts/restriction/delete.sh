if ! test $1;then
  echo "need id"
  exit 1
fi

curl -XDELETE localhost:8001/api/v1/restrictions/$1 -v -H 'userId:6' -H'tenantId:3000' -d'
