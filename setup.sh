#!/bin/bash
# setup.sh - Docker環境のセットアップと起動

set -e

CONTAINER_NAME="linux-study"
IMAGE_NAME="linux-study"
WORKSPACE_DIR="$(pwd)"

# コンテナが既に存在するかチェック
if docker ps -a --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
    echo "既存のコンテナを起動します..."
    docker start ${CONTAINER_NAME}
    docker exec -it ${CONTAINER_NAME} bash
else
    echo "新しいコンテナを作成します..."
    
    # イメージが存在しなければビルド
    if ! docker images --format '{{.Repository}}' | grep -q "^${IMAGE_NAME}$"; then
        echo "イメージをビルドしています..."
        docker build -t ${IMAGE_NAME} .
    fi
    
    # コンテナを作成して起動
    docker run -it --name ${CONTAINER_NAME} \
        -v ${WORKSPACE_DIR}:/workspace \
        ${IMAGE_NAME} bash
fi
