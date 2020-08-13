if ! test $1;then
  echo need id
fi

curl -XDELETE 10.160.32.24:30800/api/v1/apis/$1/plugins -H 'content-type:application/json'   -H'tenantId:74a2512097335196b3040bed704c65c1'
