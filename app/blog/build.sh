#!/usr/bin/env bash
set -ex
cd "$(dirname ${BASH_SOURCE[0]})"

zola build -o ./dist
