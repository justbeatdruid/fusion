if ! test $1;then
  echo "need id"
  exit 1
fi

curl -XDELETE localhost:8001/api/v1/datasource/$1/delete -v
