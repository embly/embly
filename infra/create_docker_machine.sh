set -Eeuxo pipefail

cd "$(dirname ${BASH_SOURCE[0]})"

export AWS_PROFILE=max

 docker-machine create \
    --driver amazonec2 \
    --amazonec2-subnet-id=subnet-0d00ea430d5ddc629 \
    --amazonec2-zone=c \
    --amazonec2-instance-type=m5.large \
    --amazonec2-vpc-id=vpc-051e9cc9b25711a5d \
    embly