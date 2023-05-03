#!/bin/bash

echo "Deleting stack..."
aws cloudformation delete-stack \
    --stack-name NrEntityTagSync-Stack \
    --output table \
    --no-cli-pager \
    --color on

echo "Waiting for stack delete to complete..."
aws cloudformation wait stack-delete-complete \
    --stack-name NrEntityTagSync-Stack \
    --output table \
    --no-cli-pager \
    --color on

echo "Done."
