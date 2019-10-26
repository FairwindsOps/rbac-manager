#!/bin/bash



printf "\n\n"
echo "**************************"
echo "** Begin E2E Test Setup **"
echo "**************************"
printf "\n\n"

set -e


printf "\n\n"
echo "********************************************************************"
echo "** Install rbac-manager at $CI_SHA1 **"
echo "********************************************************************"
printf "\n\n"

kubectl apply -f deploy/all.yaml
kubectl -n rbac-manager wait deployment/rbac-manager --timeout=120s --for condition=available


printf "\n\n"
echo "********************************************************************"
echo "** Test rbacDefinition **"
echo "********************************************************************"
printf "\n\n"
kubectl create serviceaccount -n rbac-manager test-rbac-manager
kubectl create clusterrole test-rbac-manager --verb="*" --resource=deployment

cat <<EOF | kubectl create -f -
apiVersion: rbacmanager.reactiveops.io/v1beta1
kind: RBACDefinition
metadata:
  name: rbac-manager-definition
rbacBindings:
  - name: admins
    subjects:
      - kind: ServiceAccount
        name: test-rbac-manager
        namespace: rbac-manager
    clusterRoleBindings:
      - clusterRole: test-rbac-manager
EOF

kubectl auth can-i create deployments --as=system:serviceaccount:rbac-manager:test-rbac-manager
