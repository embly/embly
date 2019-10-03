set -Eeuxo pipefail

cd "$(dirname ${BASH_SOURCE[0]})"

cd transport 
go generate
cd ../example
go generate
