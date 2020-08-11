if ! test $1;then
  echo "need id"
  exit 1
fi

curl localhost:8001/api/v1/apigroups/$1 -v \
 -H'userId:1' -H'tenantId:74a2512097335196b3040bed704c65c1'
