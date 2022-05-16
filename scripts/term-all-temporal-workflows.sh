#!/bin/bash

set -xeuo pipefail

tctl wf  l --ps 10000 --op --pjson | jq -r '.[].execution.workflowId' | xargs -L 1 tctl wf term --workflow_id
