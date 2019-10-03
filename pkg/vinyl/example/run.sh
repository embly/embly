set -Eeuxo pipefail

cd "$(dirname ${BASH_SOURCE[0]})"

# https://github.com/golang/go/issues/24573
go test -v -count=1 .
