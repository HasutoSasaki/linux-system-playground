# strace でシステムコールを比較するサンプル

各言語（C / Go / Rust / Python / Node.js）で Hello World を書き、strace でシステムコールを比較するための環境です。

## colima を使用している場合

colima を起動

```bash
colima start
```

## セットアップ

```bash
# イメージをビルド
docker build -t linux-study .

# コンテナを起動（カレントディレクトリをマウント）
docker run -it -v $(pwd):/workspace linux-study

# ポートを指定
docker run -it -p 8080:8080 -v $(pwd):/workspace linux-study
```

## Docker 操作

```bash
# コンテナから抜ける（コンテナは停止）
exit

# 起動中のコンテナを確認
docker ps

# 停止中のコンテナも含めて確認
docker ps -a

# 停止中のコンテナに再度入る
docker start -ai <コンテナID>

# 起動中のコンテナに入る（別ターミナルから）
docker exec -it <コンテナID> bash

# コンテナを削除
docker rm <コンテナID>
```

## ビルド

```bash
# C
gcc c/hello.c -o bin/hello_c

# Go
go build -o bin/hello_go go/hello.go

# Rust
rustc rust/hello.rs -o bin/hello_rust
```

Python と Node.js はビルド不要です。

## strace で比較

### 統計データを見る

```bash
strace -c ./bin/hello_c > /dev/null
strace -c ./bin/hello_go > /dev/null
strace -c ./bin/hello_rust > /dev/null
strace -c python3 python/hello.py > /dev/null
strace -c node node/hello.js > /dev/null
```

### write システムコールを見る

```bash
strace -e write ./bin/hello_c > /dev/null
strace -e write ./bin/hello_go > /dev/null
strace -e write ./bin/hello_rust > /dev/null
strace -e write python3 python/hello.py > /dev/null
strace -e write node node/hello.js > /dev/null
```

## 環境

| 項目    | バージョン         |
| ------- | ------------------ |
| OS      | Ubuntu 24.04       |
| Go      | 1.24.0             |
| Rust    | 最新（rustup経由） |
| Python  | 3.14               |
| Node.js | 24.x               |
