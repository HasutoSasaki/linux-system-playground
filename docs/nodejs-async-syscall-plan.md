# Node.js 非同期処理のシステムコール観察 - アクションプラン

## 目的

Node.js の非同期処理（イベントループ / libuv）が、OS レベルでどのようなシステムコールに変換されるかを strace で観察し、仕組みを理解する。

---

## 背景知識

```
┌─────────────────────────────────────────────┐
│              JavaScript コード                │
├─────────────────────────────────────────────┤
│              Node.js API                     │
│   (fs, net, http, timers, dns, ...)         │
├─────────────────────────────────────────────┤
│              libuv (イベントループ)            │
│   ┌──────────────┐  ┌───────────────────┐   │
│   │ Thread Pool   │  │ epoll/kqueue      │   │
│   │ (fs, dns,    │  │ (network I/O,     │   │
│   │  crypto)     │  │  pipes, timers)   │   │
│   └──────────────┘  └───────────────────┘   │
├─────────────────────────────────────────────┤
│              Linux Kernel                    │
│   (epoll_wait, read, write, futex, ...)     │
└─────────────────────────────────────────────┘
```

- **epoll** (Linux) / **kqueue** (macOS): I/O 多重化の仕組み。Node.js はこれで非同期 I/O を実現
- **Thread Pool**: ファイル I/O や DNS 解決など、非同期 API が存在しない操作はワーカースレッドで実行
- **futex**: スレッド間の同期に使用されるシステムコール

---

## 実験計画

### Phase 1: イベントループの基本動作を観察

#### 1-1. 空のイベントループ

**目的**: Node.js が何もしない状態でも発行するシステムコールを把握する（ベースライン）

```js
// node/async/01_baseline.js
// 何もしない（起動→終了のシステムコールを観察）
```

```bash
strace -c node node/async/01_baseline.js > /dev/null
strace -f -e trace=epoll_create1,epoll_ctl,epoll_wait node node/async/01_baseline.js > /dev/null
```

**観察ポイント**:
- `epoll_create1`: epoll インスタンスの作成
- `epoll_ctl`: ファイルディスクリプタの登録
- `epoll_wait`: イベント待ち
- `clone` / `clone3`: libuv のスレッドプール生成

#### 1-2. setTimeout による遅延

**目的**: タイマーがシステムコールレベルでどう実現されるか観察

```js
// node/async/02_timer.js
setTimeout(() => {
  console.log("timer fired");
}, 1000);
```

```bash
strace -f -T -e trace=epoll_wait,write,clock_gettime node node/async/02_timer.js
```

**観察ポイント**:
- `epoll_wait` の timeout 引数がタイマーの残り時間に対応するか
- `clock_gettime` でカーネルから時刻を取得しているか

---

### Phase 2: ファイル I/O（Thread Pool 経由）

#### 2-1. 非同期ファイル読み込み（fs.readFile）

**目的**: 非同期ファイル I/O がスレッドプールで実行されることを確認

```js
// node/async/03_file_read_async.js
const fs = require('fs');
fs.readFile('/etc/hostname', 'utf8', (err, data) => {
  console.log(data);
});
```

```bash
strace -f -T -e trace=openat,read,write,close,futex,clone3,epoll_wait node node/async/03_file_read_async.js
```

**観察ポイント**:
- `clone3`: ワーカースレッドの生成
- `openat` + `read` が **メインスレッド以外の TID** で発生するか
- `futex`: スレッド間の通知（ワーカー→メインスレッド）
- `epoll_wait` でメインスレッドがブロックし、完了通知を受け取る流れ

#### 2-2. 同期ファイル読み込み（fs.readFileSync）との比較

**目的**: 同期版との違いをシステムコールレベルで比較

```js
// node/async/04_file_read_sync.js
const fs = require('fs');
const data = fs.readFileSync('/etc/hostname', 'utf8');
console.log(data);
```

```bash
strace -f -T -e trace=openat,read,write,close,futex node node/async/04_file_read_sync.js
```

**観察ポイント**:
- `openat` + `read` が **メインスレッド（TID = PID）** で直接実行されるか
- `clone3` / `futex` が発生しないことを確認（スレッドプール不使用）

---

### Phase 3: ネットワーク I/O（epoll 直接）

#### 3-1. HTTP サーバー

**目的**: ネットワーク I/O が epoll で直接多重化されることを観察

```js
// node/async/05_http_server.js
const http = require('http');
const server = http.createServer((req, res) => {
  res.writeHead(200);
  res.end('hello\n');
});
server.listen(3000, () => {
  console.log('listening on :3000');
});
```

```bash
# ターミナル1: サーバー起動
strace -f -T -e trace=socket,bind,listen,accept4,epoll_ctl,epoll_wait,read,write,close \
  node node/async/05_http_server.js

# ターミナル2: リクエスト送信
curl http://localhost:3000
```

**観察ポイント**:
- `socket` → `bind` → `listen`: サーバーソケットの作成
- `epoll_ctl(ADD, listenfd)`: listen ソケットを epoll に登録
- `epoll_wait`: 接続待ち（ここでブロック）
- `accept4`: 新しい接続の受け入れ
- `read` → `write`: HTTP リクエスト/レスポンスのやりとり
- 全てが**メインスレッド**で実行されること（ネットワーク I/O はスレッドプール不使用）

#### 3-2. 複数の同時接続

**目的**: epoll による I/O 多重化を実際に観察

```bash
# 複数のリクエストを同時送信
for i in $(seq 1 5); do curl http://localhost:3000 & done
```

**観察ポイント**:
- 1 つの `epoll_wait` から複数の fd のイベントが返る様子
- `accept4` が連続して呼ばれる様子

---

### Phase 4: DNS 解決（Thread Pool 経由）

#### 4-1. dns.lookup（libuv Thread Pool）

**目的**: DNS 解決がスレッドプールで実行されることを確認

```js
// node/async/06_dns_lookup.js
const dns = require('dns');
dns.lookup('example.com', (err, address) => {
  console.log(address);
});
```

```bash
strace -f -T -e trace=socket,connect,sendto,recvfrom,futex,clone3,epoll_wait \
  node node/async/06_dns_lookup.js
```

**観察ポイント**:
- `clone3`: ワーカースレッドで DNS 解決
- `socket(AF_INET, SOCK_DGRAM)` → `sendto` → `recvfrom`: UDP での DNS クエリ
- ワーカースレッドの TID で実行されることを確認

#### 4-2. dns.resolve（c-ares ライブラリ）

**目的**: `dns.resolve` は c-ares を使い、epoll で直接非同期処理されることを確認

```js
// node/async/07_dns_resolve.js
const dns = require('dns');
dns.resolve('example.com', (err, addresses) => {
  console.log(addresses);
});
```

```bash
strace -f -T -e trace=socket,connect,sendto,recvfrom,epoll_ctl,epoll_wait \
  node node/async/07_dns_resolve.js
```

**観察ポイント**:
- スレッドプールを使わず、メインスレッドの epoll で DNS ソケットを監視
- `dns.lookup` との実装の違いをシステムコールレベルで確認

---

### Phase 5: Promise / async-await の観察

#### 5-1. Promise チェーン

**目的**: Promise はユーザーランドの仕組みであり、追加のシステムコールを発生させないことを確認

```js
// node/async/08_promise.js
const fs = require('fs').promises;

async function main() {
  const data = await fs.readFile('/etc/hostname', 'utf8');
  console.log(data);
  const data2 = await fs.readFile('/etc/hostname', 'utf8');
  console.log(data2);
}
main();
```

```bash
strace -f -c node node/async/08_promise.js > /dev/null
```

**観察ポイント**:
- `async/await` 自体は追加のシステムコールを発生させない
- 裏側のファイル I/O は Phase 2 と同じスレッドプール経由

---

### Phase 6: 並行処理パターンの比較

#### 6-1. 逐次 vs 並行ファイル読み込み

**目的**: `Promise.all` による並行処理がシステムコールの発行パターンにどう影響するか

```js
// node/async/09_sequential.js（逐次）
const fs = require('fs').promises;
async function main() {
  const a = await fs.readFile('/etc/hostname', 'utf8');
  const b = await fs.readFile('/etc/resolv.conf', 'utf8');
  const c = await fs.readFile('/etc/os-release', 'utf8');
  console.log(a, b, c);
}
main();
```

```js
// node/async/10_concurrent.js（並行）
const fs = require('fs').promises;
async function main() {
  const [a, b, c] = await Promise.all([
    fs.readFile('/etc/hostname', 'utf8'),
    fs.readFile('/etc/resolv.conf', 'utf8'),
    fs.readFile('/etc/os-release', 'utf8'),
  ]);
  console.log(a, b, c);
}
main();
```

```bash
# 時間計測付きで比較
strace -f -T -c node node/async/09_sequential.js > /dev/null
strace -f -T -c node node/async/10_concurrent.js > /dev/null
```

**観察ポイント**:
- 逐次: `openat` → `read` → `close` が 1 ファイルずつ直列に発生
- 並行: 複数のワーカースレッドで `openat` → `read` が同時に発生
- `futex` の呼び出し回数の違い

---

## 実行環境

既存の Docker 環境をそのまま使用:

```bash
docker build -t linux-study .
docker run -it --rm -v $(pwd):/workspace linux-study
```

## ディレクトリ構成（追加分）

```
node/
├── hello.js          # 既存
└── async/
    ├── 01_baseline.js
    ├── 02_timer.js
    ├── 03_file_read_async.js
    ├── 04_file_read_sync.js
    ├── 05_http_server.js
    ├── 06_dns_lookup.js
    ├── 07_dns_resolve.js
    ├── 08_promise.js
    ├── 09_sequential.js
    └── 10_concurrent.js
```

## strace 主要オプション早見表

| オプション | 意味 |
|-----------|------|
| `-f` | 子スレッド/子プロセスも追跡（Thread Pool 観察に必須） |
| `-T` | 各システムコールの所要時間を表示 |
| `-t` / `-tt` | タイムスタンプ付き表示 |
| `-c` | 統計サマリ表示 |
| `-e trace=...` | 特定のシステムコールのみフィルタ |
| `-p PID` | 実行中のプロセスにアタッチ |
| `-o FILE` | 出力をファイルに保存 |

## 注目すべきシステムコール一覧

| システムコール | 役割 | Node.js での登場場面 |
|--------------|------|---------------------|
| `epoll_create1` | epoll インスタンス作成 | Node.js 起動時 |
| `epoll_ctl` | fd の監視登録/削除 | ソケット、パイプ、タイマーの登録 |
| `epoll_wait` | イベント待ち | イベントループの心臓部 |
| `clone3` | スレッド/プロセス生成 | libuv Thread Pool の初期化 |
| `futex` | スレッド同期 | Thread Pool ↔ メインスレッドの通知 |
| `openat` | ファイルオープン | fs 操作 |
| `read` / `write` | データ読み書き | ファイル・ネットワーク I/O |
| `socket` | ソケット作成 | ネットワーク・DNS |
| `accept4` | 接続受け入れ | HTTP サーバー |
| `clock_gettime` | 時刻取得 | タイマー管理 |

## 実行順序

1. **Phase 1** → ベースラインを把握（他のフェーズの比較基準になる）
2. **Phase 2** → ファイル I/O でスレッドプールの動作を理解
3. **Phase 3** → ネットワーク I/O で epoll の動作を理解
4. **Phase 4** → DNS で「同じ非同期でも実装が違う」ことを理解
5. **Phase 5** → Promise/async-await はシステムコールに影響しないことを確認
6. **Phase 6** → 逐次 vs 並行の実際のシステムコール差を体感

## 検証結果の記録

各フェーズの結果は `results/` ディレクトリに保存:

```bash
# 例
strace -f -c node node/async/03_file_read_async.js > /dev/null 2> results/phase2_async_read.txt
strace -f -c node node/async/04_file_read_sync.js > /dev/null 2> results/phase2_sync_read.txt
```
