#!/bin/bash

source $(dirname "$0")/init.sh

println "\n%s" "-- BUILD -----------------------------------------------------------------------"
println "Root directory:                          $ROOT_DIR"
println "Dist directory:                          $DIST_DIR"
println "Stage directory:                         $STAGE_DIR"
println "%s\n" "--------------------------------------------------------------------------------"

println "Building lambda zip package..."
build_zip
