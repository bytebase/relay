#!/bin/sh
# ===========================================================================
# File: build_docker.sh
# Description: usage: ./build_docker.sh
# ===========================================================================

# exit when any command fails
set -e

echo "Start building Webhook Relay docker image ..."

docker build -f ./Dockerfile\
    --build-arg GO_VERSION="$(go version)" \
    --build-arg GIT_COMMIT="$(git rev-parse HEAD)"\
    --build-arg BUILD_TIME="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"  \
    --build-arg BUILD_USER="$(id -u -n)" \
    -t bytebase/relay .

echo "${GREEN}Completed building Bytebase Webhook Relay docker image.${NC}"
echo ""
echo "Command to tag and push the image"
echo ""
echo "$ docker tag bytebase/relay latest; docker push bytebase/relay:latest"
echo ""
echo "Command to start Bytebase Webhook Relay on http://localhost:8080"
echo ""
echo "docker run --init --name relay --restart always --publish 8080:2830 bytebase/relay --ref-prefix=refs/heads/release/ --lark-url=https://open.feishu.cn/open-apis/bot/v2/hook/xxxxxxxxxxxxxxxxx"
echo ""
