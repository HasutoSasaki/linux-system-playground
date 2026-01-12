#!/bin/bash
# clean.sh - Docker環境のクリーンアップ

set -e

CONTAINER_NAME="linux-study"
IMAGE_NAME="linux-study"

echo "コンテナとイメージを削除します..."

# コンテナを停止・削除
if docker ps -a --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
    docker stop ${CONTAINER_NAME} 2>/dev/null || true
    docker rm ${CONTAINER_NAME}
    echo "コンテナを削除しました"
fi

# イメージを削除
if docker images --format '{{.Repository}}' | grep -q "^${IMAGE_NAME}$"; then
    docker rmi ${IMAGE_NAME}
    echo "イメージを削除しました"
fi

echo "クリーンアップ完了"
