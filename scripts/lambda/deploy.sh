#!/bin/bash

SCRIPT_DIR="$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
ROOT_DIR="$(dirname $(dirname ${SCRIPT_DIR}))"

echo "Deploying stack..."
aws cloudformation deploy \
    --stack-name NrEntityTagSync-Stack \
    --template-file $ROOT_DIR/deployments/cf-template.yaml \
    --output table \
    --no-cli-pager \
    --color on \
    --parameter-overrides file://$ROOT_DIR/deployments/cf-params.json

echo "Done."
