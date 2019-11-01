#!/bin/bash

set -e

wget -O /usr/local/bin/yq "https://github.com/mikefarah/yq/releases/download/2.4.0/yq_linux_amd64"
chmod +x /usr/local/bin/yq

if [ -z "$CI_SHA1" ]; then
    echo "CI_SHA1 not set. Something is wrong"
    exit 1
else
    echo "CI_SHA1: $CI_SHA1"
fi

yq w -d5 -i deploy/all.yaml 'spec.template.spec.containers[0].image' "quay.io/reactiveops/rbac-manager:dev-$CI_SHA1"
cat deploy/all.yaml

docker cp deploy e2e-command-runner:/
