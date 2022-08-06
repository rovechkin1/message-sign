

# some issues with  ssl, fix it later
# helm install mongo bitnami/mongodb --set auth.rootPassword="aaaaaa123#"
# disable auth for now
helm install mongo bitnami/mongodb --set auth.enabled=false