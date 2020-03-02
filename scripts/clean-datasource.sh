#!/bin/bash

kubectl get datasources -o go-template --template='{{range .items}}{{.metadata.name}} {{.spec.type}}
{{end}}'|grep "datawarehouse"|awk '{print $1}'|xargs --no-run-if-empty kubectl delete datasource
