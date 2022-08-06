
# creates k8s secret from keys.csv
kubectl create secret generic sign-keys --from-file=keys.csv