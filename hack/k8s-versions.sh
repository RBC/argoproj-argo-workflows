#!/bin/bash

# Centralized config to define the minimum and maximum tested Kubernetes versions.
# This is used in the CI workflow for e2e tests, the devcontainer, and to generate docs.
declare -A K8S_VERSIONS=(
  [min]=v1.31.9
  [max]=v1.33.1
)
