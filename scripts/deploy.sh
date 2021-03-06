#!/bin/bash

set -x

tf=$(mktemp)
cat <<EOF>>${tf}
---
apiVersion: v1
kind: Service
metadata:
  name: fusion-apiserver
  namespace: default
  labels:
    app: fusion-apiserver
    version: v0.1-alpha
spec:
  type: ClusterIP
  ports:
    - name: service
      port: 8001
      protocol: TCP
      targetPort: 8001
  selector:
    app: fusion-apiserver
    version: v0.1-alpha
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: fusion-apiserver
    version: v0.1-alpha
  name: fusion-apiserver
  namespace: default
spec:
  replicas: 2
  selector:
    matchLabels:
      app: fusion-apiserver
      version: v0.1-alpha
  template:
    metadata:
      labels:
        app: fusion-apiserver
        version: v0.1-alpha
    spec:
      affinity:
        podAntiAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
          - labelSelector:
              matchExpressions:
              - key: app
                operator: In
                values:
                - fusion-apiserver
              - key: version
                operator: In
                values:
                - v0.1-alpha
            topologyKey: "kubernetes.io/hostname"
      serviceAccount: fusion-apiserver
      containers:
      - command:
        - fusion-apiserver
        - --v=5
        env:
        - name: MY_POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        image: registry.cmcc.com/library/fusion-apiserver:0.1.0
        imagePullPolicy: Always
        name: fusion-apiserver
        ports:
        - containerPort: 8001
          name: http
          protocol: TCP
        #resources:
        #  requests:
        #    cpu: 0.1
        #    memory: 100Mi
        volumeMounts:
        - mountPath: /data
          name: data
      restartPolicy: Always
      securityContext:
        fsGroup: 0
        runAsUser: 0
      volumes:
      - emptyDir: {}
        name: data
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: fusion-apiserver
  labels:
    app: fusion-apiserver
    version: v0.1-alpha
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
subjects:
- kind: ServiceAccount
  name: fusion-apiserver
  namespace: default
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: fusion-apiserver
  namespace: default
  labels:
    app: fusion-apiserver
    version: v0.1-alpha
EOF
kubectl delete -f ${tf}
kubectl create -f ${tf}

rm -rf ${tf}
