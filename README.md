# Docker Ubuntu 環境の管理

## 初回セットアップ

```bash
# Ubuntu 20.04コンテナを作成して起動
docker run -it --name linux-study ubuntu:20.04 bash

# コンテナ内で必要なパッケージをインストール
apt update
apt install -y strace vim curl wget git
```

## 日常的な操作

### コンテナの起動・停止

```bash
# コンテナを起動
docker start linux-study

# コンテナに接続
docker exec -it linux-study bash

# コンテナから抜ける（コンテナは起動したまま）
exit

# コンテナを停止
docker stop linux-study
```

### コンテナの状態確認

```bash
# 起動中のコンテナ一覧
docker ps

# すべてのコンテナ一覧（停止中も含む）
docker ps -a

# コンテナの詳細情報
docker inspect linux-study
```

## ファイルのやり取り

```bash
# ホスト → コンテナ
docker cp ./local-file.txt linux-study:/root/

# コンテナ → ホスト
docker cp linux-study:/root/remote-file.txt ./
```

## クリーンアップ

```bash
# コンテナを削除（停止してから）
docker stop linux-study
docker rm linux-study

# イメージも削除する場合
docker rmi ubuntu:20.04
```

## 再作成する場合

```bash
# 古いコンテナを削除
docker rm -f linux-study

# 新しいコンテナを作成
docker run -it --name linux-study ubuntu:20.04 bash
```

## Tips

- コンテナ内の変更は保持されるので、一度インストールしたパッケージは再起動後も使える
- コンテナを削除すると、内部のデータはすべて消える
- 重要なデータは `docker cp` でホストに保存しておく
