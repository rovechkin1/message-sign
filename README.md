# Message Signer

## Local development
```
# build service
make build-service
```

```
# build key generator utility, writes keys to a file keys.csv
make build-key-gen
```

```
# build record generator which writes records to mongo db
make build-record-gen
```

## Setup
```
# generate keys and save them in keys.csv
bin/key-generator

# start mongo db
./start-local-mongo.sh

# populate mongodb with records
bin/record-generator

# start service
bin/service

```

## API
```
# sign records /sign/<batch_size>
curl localhost:8080/sign/30000

# get stats /stats
curl localhost:8080/stats
ruslans-MBP:message-sign ruslan$ curl localhost:8080/stats
stats: {"total_records":1000,"signed_records":800,"unsigned_records":200}

```
# Kubernetes

## Build service container
Build
```
./build-docker.sh

```

Push to gcr
```
gcloud auth configure-docker
./push-to-gcr.sh
```

## MongoDb

```
# install Helm repo
helm repo add bitnami https://charts.bitnami.com/bitnami
```

```
# deploy mongodb into k8s cluster
./k8s-deploy-mongodb.sh

```

To Connect
```
kubectl run --namespace default mongo-mongodb-client --rm --tty -i --restart='Never'  --image docker.io/bitnami/mongodb:6.0.0-debian-11-r0 --command -- bash

mongosh admin --host "mongo-mongodb" --authenticationDatabase admin 
```

## Deploy signing service
Create k8s secret from keys file
```
./k8s-create-secret.sh
```
```
./k8s-deploy-msg-signer.sh
```

## Sign records
Connect to a signer pod
```
kubectl get pods | grep msg-signer
msg-signer-78d88cfb86-td9f2      1/1     Running   0          54s

kubectl exec -it msg-signer-78d88cfb86-td9f2 -- bash
```
Generate records
```
./record-generator 1000
```


## Uninstall

```
helm uninstall msg-signer 
helm uninstall mongo
```

