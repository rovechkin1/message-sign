

# some issues with  ssl to connect to mongodb with auth
# disable it, access to mongodb is only possible from within cluster
# helm install record-generator bitnami/mongodb --set auth.rootPassword="aaaaaa123#"
helm install mongo bitnami/mongodb --set auth.enabled=false