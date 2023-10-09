# wait up to 2 minutes for rbac-manager to create the binding
counter=0
error=$((0))
until kubectl get -n rbac-manager serviceaccount/test-rbac-manager; do
  let "counter=counter+1"
  sleep 10
  if [ $counter -gt 11 ]; then
    break
  fi
done

kubectl get -n rbac-manager serviceaccount/test-rbac-manager
error=$(( error | $? ))
if [ "$error" -eq 1 ]; then
    >&2 echo "error: The Service account must exists"
fi
kubectl delete -n rbac-manager serviceaccount/test-rbac-manager
kubectl get -n rbac-manager serviceaccount/test-rbac-manager
error=$(( error | $? ))
if [ "$error" -eq 1 ]; then
    >&2 echo "error: The Service account must be recreated"
fi

# ImagePullSecret is created
contents=$(kubectl get -n rbac-manager serviceaccount/test-rbac-manager -oyaml | yq 'select(.imagePullSecrets[] | .name == "robot-secret")')
if [ -z "$contents" ]; then
  error=$(( error | 1 ))
fi
if [ "$error" -eq 1 ]; then
    >&2 echo "error: ImagePullSecret \"robot-secret\" must exists"
fi

# ImagePullSecret is re-created if deleted
cat <<EOF | kubectl patch -n rbac-manager serviceaccount/test-rbac-manager --type=merge -p "$(cat -)"
{
  "imagePullSecrets": []
}
EOF
contents=$(kubectl get -n rbac-manager serviceaccount/test-rbac-manager -oyaml | yq 'select(.imagePullSecrets[] | .name == "robot-secret")')
if [ -z "$contents" ]; then
  error=$(( error | 1 ))
fi
if [ "$error" -eq 1 ]; then
    >&2 echo "error: ImagePullSecret \"robot-secret\" must be re-created"
fi

# If ImagePullSecret is added it should not be removed

cat <<EOF | kubectl patch -n rbac-manager serviceaccount/test-rbac-manager --type=json -p "$(cat -)"
[
  {
    "op": "add",
    "path": "/imagePullSecrets/-",
    "value": {
      "name": "new-secret-name"
    }
  }
]
EOF
contents=$(kubectl get -n rbac-manager serviceaccount/test-rbac-manager -oyaml | yq 'select(.imagePullSecrets[] | .name == "new-secret-name")')
if [ -z "$contents" ]; then
  error=$(( error | 1 ))
fi
if [ "$error" -eq 1 ]; then
    >&2 echo "error: ImagePullSecret \"new-secret-name\" must be kept"
fi

exit $error