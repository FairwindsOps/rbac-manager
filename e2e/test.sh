#!/bin/bash

BASE_DIR=$(dirname $BASH_SOURCE)

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

kubectl apply -f /deploy/
kubectl -n rbac-manager wait deployment/rbac-manager --timeout=120s --for condition=available

printf "\n\n"
echo "********************************************************************"
echo "** Install and run Chainsaw **"
echo "********************************************************************"
printf "\n\n"

cd "/chainsaw"

curl -sL https://github.com/kyverno/chainsaw/releases/download/v0.2.14/chainsaw_linux_amd64.tar.gz -o linux_amd64.tar.gz
tar -xvf linux_amd64.tar.gz chainsaw
rm linux_amd64.tar.gz
chmod +x chainsaw

./chainsaw test

if [ $? -ne 0 ]; then
  exit 1
fi
