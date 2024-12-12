#!/bin/bash

source $(dirname "$0")/init.sh

AWS_STACK_NAME=${AWS_STACK_NAME:-"$APP_DIR_NAME"}
AWS_CF_DELETE_OPTS=${AWS_CF_DELETE_OPTS:-""}

println "\n%s" "-- DELETE -----------------------------------------------------------------------"
println "Root directory:                          $ROOT_DIR"
println "Dist directory:                          $DIST_DIR"
println "Stage directory:                         $STAGE_DIR"
println "AWS region:                              $AWS_REGION"
println "Stack Name:                              $AWS_STACK_NAME"
println "Delete stack options:                    $AWS_CF_DELETE_OPTS"
println "%s\n" "---------------------------------------------------------------------------------"

println "Deleting stack $AWS_STACK_NAME..."
aws cloudformation delete-stack \
    --stack-name $AWS_STACK_NAME \
    --output table \
    --no-cli-pager \
    --color on \
    $AWS_CF_DELETE_OPTS

println "Waiting for stack delete to complete..."
aws cloudformation wait stack-delete-complete \
    --stack-name $AWS_STACK_NAME \
    --output table \
    --no-cli-pager \
    --color on

println "Done."
