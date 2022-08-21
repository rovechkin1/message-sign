

#kubectl run --namespace default mongo-mongodb-client --rm --tty -i --restart='Never' --image docker.io/bitnami/mongodb:6.0.0-debian-11-r0 --command -- bash

# some issues with  ssl to connect to mongodb with auth
# disable it, access to mongodb is only possible from within cluster
helm install -f mongo-values.yaml mongo bitnami/mongodb --set auth.enabled=false --set architecture="replicaset"
