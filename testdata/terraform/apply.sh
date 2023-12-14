#!/usr/bin/env bash

set -e

terraform apply \
  -var prefix=testkit- \
  -var region=$AWS_DEFAULT_REGION \
  -var vpc_id=$TESTKIT_VPC_ID

terraform show -json | jq . > terraform.show.$(date +%Y%m%d%H%M%S).json
