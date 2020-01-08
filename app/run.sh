#! /usr/bin/env bash
set -Eeuxo pipefail

cd "$(dirname ${BASH_SOURCE[0]})"

SOURCE=$(pwd)

cd ~/embly/cmd/embly && go install

cd $SOURCE
# embly run out.tar.gz
embly dev
