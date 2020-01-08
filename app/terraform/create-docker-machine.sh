set -Eeuxo pipefail

cd "$(dirname ${BASH_SOURCE[0]})"

export AWS_PROFILE=max

docker-machine create embly-build --driver amazonec2 \
    --amazonec2-instance-type=m5.large\
    --amazonec2-region=us-east-1 \
    --amazonec2-root-size=100 \
    --amazonec2-vpc-id=vpc-0b558057e543ced7a 