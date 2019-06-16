set -Eeuxo pipefail

cd "$(dirname ${BASH_SOURCE[0]})"

docker-compose build img

docker run -it \
    --privileged \
    --volume "${HOME}/.docker:/root/.docker:ro" \
    --security-opt seccomp=unconfined --security-opt apparmor=unconfined \
    internal/embly-compile-rust-wasm-img-for-cache \
    build -t user/myimage .

docker commit $(docker ps -aq | head -n 1) maxmcd/embly-img:rust