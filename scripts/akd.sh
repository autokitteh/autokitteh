#!/bin/bash

set -euo pipefail

if [[ "${PERSIST-}" == "1" ]]; then
  export AKD_DEFAULT_STORE_SPEC="gorm:sqlite:autokitteh.sqlite"
fi

export AKD_HTTP_CORS=1

export AKD_LOG_LEVEL="${LOG_LEVEL-info}"

opts=("--enable" "defaults")

export AKD_DASHBOARD_TEMPLATES_PATH="internal/app/dashboardsvc/templates"

export AKD_INIT_PATHS="${AKD_INIT_PATHS-"examples/manifests"}"

export AKD_INITD_DIR="examples/initd"

#
# TEMPORALITE
#

if [[ "${TEMPORALITE-}" == "0" ]]; then
  opts=("--disable" "temporalite")
elif [[ "${TEMPORALITE-}" == "1" ]]; then
  opts=("--enable" "temporalite")
fi

#
# CREDENTIALS STORE
#

export AKD_CREDS_STORE_TYPE="fs"
export AKD_CREDS_STORE_FS_ROOT_PATH="crypt/data/creds"

#
# FS
#

if [[ "${FS-}" == "1" ]]; then
  opts+=("--enable" "fseventsrc")
fi

#
# GITHUB
#

if [[ "${GITHUB-}" == "1" ]]; then
  opts+=("--enable" "githubeventsrcsvc" "--enable" "githubinstalls")

  source crypt/data/github-local.sh

  export AKD_DEFAULT_GITHUB_REPOS=softkitteh
  export AKD_GITHUB_INSTALLS_STORE_TYPE="fs"
  export AKD_GITHUB_INSTALLS_STORE_FS_ROOT_PATH="crypt/data/github-installations"
fi

#
# CRON
#

if [[ "${CRON-}" == "1" ]]; then
  opts+=("--enable" "croneventsrcsvc")
fi

#
# GOOGLE OAUTH
#

[[ -r crypt/data/google-oauth-local.sh ]] && source crypt/data/google-oauth-local.sh

export AKD_GOOGLE_OAUTH_REDIRECT_URL="https://autokitteh.ngrok.io/google-oauth/oauth/installed"
export AKD_GOOGLE_OAUTH_SCOPES="https://www.googleapis.com/auth/spreadsheets"

#
# SLACK
#

if [[ "${SLACK-}" == "1" ]]; then
  opts+=("--enable" "slackeventsrcsvc")

  source crypt/data/slack-local.sh

  export AKD_SLACK_EVENT_SOURCE_ID="autokitteh.slack"
  export AKD_SLACK_EVENT_SOURCE_OAUTH_ENABLED="${SLACK_OAUTH-1}"
  export AKD_SLACK_EVENT_SOURCE_SOCKET_MODE="${SLACK_SOCKET_MODE-0}"
  export AKD_SLACK_EVENT_SOURCE_DEBUG="${SLACK_DEBUG-0}"

  export AKD_DEFAULT_SLACK_TEAM_IDS=softkitteh:TFPTT3QFN,autokitteh:T02T1NWQK62
fi

#
# SECRETS
#

export AKD_SECRETS_STORE_TYPE="fs"
export AKD_SECRETS_STORE_FS_ROOT_PATH="crypt/data/secrets"

#
# Execute
#

if [[ "${DEBUG-}" == "1" ]]; then
  set -x && TEMPORAL_DEBUG=1 dlv exec ./bin/akd -- "${opts[@]}" "$@"
elif [[ "${GDEBUG-}" == "1" ]]; then
  set -x && TEMPORAL_DEBUG=1 gdlv exec ./bin/akd -- "${opts[@]}" "$@"
else
  set -x && exec ./bin/akd "${opts[@]-}" "$@"
fi
