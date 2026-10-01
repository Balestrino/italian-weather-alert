#!/usr/bin/env bash
set -euo pipefail

if [[ $# -lt 2 ]] || [[ "$1" != development && "$1" != staging ]]; then
  echo 'Usage: scripts/compose-env.sh development|staging <docker compose command>' >&2
  exit 2
fi

environment=$1
shift
repo_root=$(git -C "$(dirname "$0")" rev-parse --show-toplevel)
env_file="$repo_root/.local/$environment.env"
if [[ ! -f "$env_file" ]]; then
  echo "Missing local environment file: $env_file" >&2
  exit 1
fi

project=iwa
if [[ "$environment" == staging ]]; then
  project=iwa-staging
fi

unset_args=()
while IFS='=' read -r name _; do
  if [[ "$name" == IWA_* || "$name" == COMPOSE_* ]]; then
    unset_args+=(-u "$name")
  fi
done < <(env)

cd "$repo_root"
exec env "${unset_args[@]}" docker compose --env-file "$env_file" -p "$project" "$@"
