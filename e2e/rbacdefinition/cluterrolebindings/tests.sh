# wait up to 2 minutes for rbac-manager to create the binding
counter=0
until kubectl get clusterrolebinding/rbac-manager-definition-admins-test-rbac-manager; do
  let "counter=counter+1"
  sleep 10
  if [ $counter -gt 11 ]; then
    break
  fi
done
kubectl auth can-i create deployments --as=system:serviceaccount:rbac-manager:test-rbac-manager