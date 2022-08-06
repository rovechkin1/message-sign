# Kubernetes guide

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
./record-generator 100000
```

Sign messages
```
curl http://localhost:8080/sign/5000
```
Check progress
```

curl http://localhost:8080/stats
stats: {"signed_records":38680,"unsigned_records":10208}
```

View records
```
kubectl run --namespace default mongo-mongodb-client --rm --tty -i --restart='Never'  --image docker.io/bitnami/mongodb:6.0.0-debian-11-r0 --command -- bash

mongosh admin --host "mongo-mongodb" --authenticationDatabase admin 

use msg-signer
db.signedrecords.find()

```


## Uninstall

```
helm uninstall msg-signer 
helm uninstall mongo
```

