if ! test $1;then
  echo "need id"
  exit 1
fi

curl localhost:8001/api/v1/api/$1/get -v
