#!/usr/bin/env bash

set -e

terraform apply \
  -var prefix=testkit-tfs3-

terraform show -json | jq . > terraform.show.$(date +%Y%m%d%H%M%S).json
