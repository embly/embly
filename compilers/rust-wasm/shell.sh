set -Eeuxo pipefail

cd "$(dirname ${BASH_SOURCE[0]})"

export AWS_PROFILE=max

docker-compose down
./build.sh
docker-compose run app bash
