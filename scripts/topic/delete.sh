if ! test $1;then
  echo "need id"
  exit 1
fi

curl -XDELETE localhost:8001/api/v1/topic/$1/delete -v
