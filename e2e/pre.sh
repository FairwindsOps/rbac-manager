#!/bin/bash

set -e

wget -O /usr/local/bin/yq "https://github.com/mikefarah/yq/releases/download/v4.35.1/yq_linux_amd64"
chmod +x /usr/local/bin/yq

if [ -z "$CI_SHA1" ]; then
    echo "CI_SHA1 not set. Something is wrong"
    exit 1
else
    echo "CI_SHA1: $CI_SHA1"
fi

printf "\n\n"
echo "********************************************************************"
echo "** LOADING IMAGES TO DOCKER AND KIND **"
echo "********************************************************************"
printf "\n\n"
docker load --input /tmp/workspace/docker_save/rbac-manager_${CI_SHA1}-amd64.tar
export PATH=$(pwd)/bin-kind:$PATH
kind load docker-image --name e2e quay.io/reactiveops/rbac-manager:${CI_SHA1}-amd64
printf "\n\n"
echo "********************************************************************"
echo "** END LOADING IMAGE **"
echo "********************************************************************"
printf "\n\n"

yq -i '.spec.template.spec.containers[0].image = "quay.io/reactiveops/rbac-manager:${CI_SHA1}-amd64"' deploy/3_deployment.yaml
yq -i '.spec.template.spec.containers[0].imagePullPolicy = "IfNotPresent"' deploy/3_deployment.yaml
cat deploy/3_deployment.yaml

docker cp deploy e2e-command-runner:/
