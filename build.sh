set -Eeuxo pipefail

cd "$(dirname ${BASH_SOURCE[0]})"

cd ./control

export EMBLY_USER=$(id -u)
docker-compose build app