#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
Usage: scripts/release-image.sh build|publish

Run from a clean revision in the shared checkout. Set IWA_RELEASE_IMAGE_REPOSITORY to override
ghcr.io/balestrino/italian-weather-alert. Build first, test that exact image in
staging, then publish it. The operator approves the printed registry digest.
EOF
}

if [[ $# -ne 1 ]] || [[ "$1" != build && "$1" != publish ]]; then
  usage >&2
  exit 2
fi

command -v git >/dev/null
command -v docker >/dev/null
repo_root=$(git rev-parse --show-toplevel)
cd "$repo_root"

if [[ -n $(git status --porcelain --untracked-files=normal) ]]; then
  echo 'Release checkout has uncommitted or untracked files.' >&2
  exit 1
fi

revision=$(git rev-parse --verify HEAD)
image_repository=${IWA_RELEASE_IMAGE_REPOSITORY:-ghcr.io/balestrino/italian-weather-alert}
if [[ ! $image_repository =~ ^[a-z0-9][a-z0-9._/-]*$ ]] || [[ $image_repository != ghcr.io/* ]]; then
  echo 'IWA_RELEASE_IMAGE_REPOSITORY must be a lowercase ghcr.io image path.' >&2
  exit 2
fi
image_tag="$image_repository:$revision"

if [[ $1 == build ]]; then
  docker build --pull \
    --build-arg "VCS_REF=$revision" \
    --tag "$image_tag" .
  echo "Staging image: $image_tag"
  echo 'Set IWA_APP_IMAGE to this tag for staging checks, then run publish.'
  exit 0
fi

actual_revision=$(docker image inspect --format '{{index .Config.Labels "org.opencontainers.image.revision"}}' "$image_tag")
if [[ $actual_revision != "$revision" ]]; then
  echo 'The local image revision label does not match this checkout.' >&2
  exit 1
fi

docker push "$image_tag"
image_digest=$(docker image inspect --format '{{range .RepoDigests}}{{println .}}{{end}}' "$image_tag" | awk -v prefix="$image_repository@" 'index($0, prefix) == 1 { print; exit }')
if [[ -z $image_digest ]]; then
  echo 'Push succeeded, but Docker did not report a registry digest. Do not approve this release.' >&2
  exit 1
fi
docker pull "$image_digest"
local_id=$(docker image inspect --format '{{.Id}}' "$image_tag")
published_id=$(docker image inspect --format '{{.Id}}' "$image_digest")
if [[ $local_id != "$published_id" ]]; then
  echo 'The registry digest did not pull the tested local image. Do not approve this release.' >&2
  exit 1
fi
echo "Approved-image candidate: $image_digest"
echo 'Record this digest with staging evidence; production requires separate operator approval.'
