set -e
my_proj=$(gcloud config get project)
docker tag msg-signer:latest gcr.io/"$my_proj"/msg-signer:v1.0
docker push gcr.io/"$my_proj"/msg-signer:v1.0