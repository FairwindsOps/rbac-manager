#!/bin/bash

BASE_DIR="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"

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

kubectl apply -f deploy/
kubectl -n rbac-manager wait deployment/rbac-manager --timeout=120s --for condition=available

bash "$BASE_DIR/rbacdefinition/run.sh"
if [ $? -ne 0 ]; then
  exit 1
fi
