set -Eeuxo pipefail

cd "$(dirname ${BASH_SOURCE[0]})"

cd ./control
docker-compose down 
docker-compose run app