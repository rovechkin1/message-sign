
set -e
set -x
my_proj=$(gcloud config get project)
action="install"
if [ $1 != "" ]; then
  action="$1"
fi
helm "$action" msg-signer charts --set image.repository=gcr.io/"$my_proj"/msg-signer