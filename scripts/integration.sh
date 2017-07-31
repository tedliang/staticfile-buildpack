#!/usr/bin/env bash
set -euo pipefail
set -x

export ROOT=$(dirname $(readlink -f ${BASH_SOURCE%/*}))
if [ ! -f "$ROOT/.bin/ginkgo" ]; then
  (cd "$ROOT/src/staticfile/vendor/github.com/onsi/ginkgo/ginkgo/" && go install)
fi
if [ ! -f "$ROOT/.bin/buildpack-packager" ]; then
  (cd "$ROOT/src/staticfile/vendor/github.com/cloudfoundry/libbuildpack/packager/buildpack-packager" && go install)
fi

FILE_VERSION=$(cat VERSION | head -1 | cut -f1 -d ' ')
VERSION="$FILE_VERSION.$RANDOM"
GINKGO_NODES=${GINKGO_NODES:-3}
GINKGO_ATTEMPTS=${GINKGO_ATTEMPTS:-2}

if [ "${CACHED:-true}" = "false" ]; then
  buildpack-packager --cached=false --version=$VERSION
  cf update-buildpack staticfile_buildpack -p staticfile_buildpack-v$VERSION.zip

  cd $ROOT/src/staticfile/integration
  ginkgo -r --flakeAttempts=$GINKGO_ATTEMPTS -nodes $GINKGO_NODES -- --cached=false --version=$VERSION
else
  buildpack-packager --cached --version=$VERSION
  cf update-buildpack staticfile_buildpack -p staticfile_buildpack-cached-v$VERSION.zip

  cd $ROOT/src/staticfile/integration
  ginkgo -r --flakeAttempts=$GINKGO_ATTEMPTS -nodes $GINKGO_NODES -- --cached --version=$VERSION
fi
