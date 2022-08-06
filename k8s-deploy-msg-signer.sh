
set -e
set -x
my_proj=$(gcloud config get project)
helm install msg-signer charts --set image.repository=gcr.io/"$my_proj"/msg-signer