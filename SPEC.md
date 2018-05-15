Specifications
==============

## IPアドレスの生成ルール

sabakanの設定は引数以外にも用意されている。etcdの`<prefix>/config`に保存されており、etcd memberから常に参照可能である。
これらはIPアドレス算出などに使用される。機材が登録された後に変更をすると**登録された機材のIPアドレスの全てに影響があるため、一つでも機材が登録されている場合は変更できない。**

IPアドレスの算出式 ([IPAM設定](#prefixconfig) と [機材情報](#prefixmachinesserial) をもとに算出をおこなう)

```
node0 = INET_NTOA(INET_ATON(config.NodeIPv4Offset) + 2^(config.NodeRackShift) * config.NodeIPPerNode * machine.Rack + machine.IndexInRack)
node1 = INET_NTOA(INET_ATON(config.NodeIPv4Offset) + 2^(config.NodeRackShift) * config.NodeIPPerNode * machine.Rack + machine.IndexInRack + 2^(config.NodeRackShift))
node2 = INET_NTOA(INET_ATON(config.NodeIPv4Offset) + 2^(config.NodeRackShift) * config.NodeIPPerNode * machine.Rack + machine.IndexInRack + 2^(config.NodeRackShift + 1))
BMC   = INET_NTOA(INET_ATON(config.BMCIPv4Offset) + 2^(config.BMCRackShift) * config.BMCIPPerNode * machine.Rack + machine.IndexInRack)
```

例:
```
IPAMConfig:
  NodeIPv4Offset:      10.69.0.0/26
  NodeRackShift:       6
  NodeIPPerNode:       3
  BMCIPv4Offset:       10.72.17.0/27
  BMCRachiShift:       5
  BMCIPPerNode:        1

WorkerNode1:
  Rack:                0
  IndexInRack:         4
  割り当てられるIPアドレス:
    node0:             10.69.0.4
    node1:             10.69.0.68
    node2:             10.69.0.132
    bmc:               10.72.17.4

WorkerNode2:
  Rack:                1
  IndexInRack:         5
  割り当てられるIPアドレス:
    node0:             10.69.0.197
    node1:             10.69.0.5
    node2:             10.69.0.69
    bmc:               10.72.17.37
```

## REST API

### 共通仕様

- JSONパース失敗時: 400 Bad Request
- 内部エラー: 500 Internal Server Error

### `PUT /api/v1/config`

sabakan用設定パラメータの追加。IPアドレス算出などに使用される。**機材が登録された後に変更をすると登録されたIPアドレスの全てに影響があるため、一つでも機材が登録されている場合は変更できない。**

**成功時のレスポンス**

- ステータスコード: 200 OK

**失敗ケース**

- 既に機材が登録されている。 (500 Internal Server Error)

```console
$ curl -XPUT localhost:8888/api/v1/config -d '
{
   "max-nodes-in-rack": 28,
   "node-ipv4-offset": "10.69.0.0/26",
   "node-rack-shift": 6,
   "node-index-offset": 3,
   "bmc-ipv4-offset": "10.72.17.0/27",
   "bmc-rack-shift": 5,
   "node-ip-per-node": 3,
   "bmc-ip-per-node": 1
}'
```

### `GET /api/v1/config`

sabakan用設定パラメータの取得。

**成功時のレスポンス**

- ステータスコード: 200 OK
- レスポンスボディ: `application/json` 設定のJSON

**失敗ケース**

- etcdに `/<prefix>/config` が存在しない(404 Not Found)

```console
$ curl -XGET localhost:8888/api/v1/config
{
   "max-nodes-in-rack": 28,
   "node-ipv4-offset": "10.69.0.0/26",
   "node-rack-shift": 6,
   "node-index-offset": 3,
   "bmc-ipv4-offset": "10.72.17.0/27",
   "bmc-rack-shift": 5,
   "node-ip-per-node": 3,
   "bmc-ip-per-node": 1
}
```

### `POST /api/v1/machines`

機材エントリーの追加。`status`は`running`にセットされ、`index-in-rack`(ラック内でのノードのインデックス) および、ホストのIPアドレスとBMCのIPアドレスを自動で割り当てる。
全ての機材の追加はatomicに行われ、一つでも機材の追加に失敗すると、全て失敗する。つまり全機材が登録されていないか、されている状態のみで、一部の機材が登録されるということはない。

リクエストのボディには、以下の機材情報のリストをJSON形式で指定する。

Field                        | Description
-----                        | -----------
`serial=<serial>`            | 機材のシリアル番号
`datacenter=<datacenter>`    | 機材が置かれているデータセンター
`rack=<rack>`                | 機材が置かれているラック番号(省略した場合は `0` が割り当てられる)
`role=<role>`                | 機材のロール(`boot` or `worker`)
`product=<product>`          | 機材の製品名(`R630` 等)

**成功時のレスポンス**

- ステータスコード: 201 Created

**失敗ケース**

- 既に同じシリアルの機材が登録されている: 409 Conflict
- 指定したラックにすでにブートサーバ用の機材が登録されている: 409 Conflict

```console
$ curl -i -X POST \
   -H "Content-Type:application/json" \
   -d \
'[{
  "serial": "1234abcd",
  "product": "R630",
  "datacenter": "ty3",
  "rack": 1,
  "role": "boot"
}]' \
 'http://localhost:8888/api/v1/machines'
```

### `GET /api/v1/machines`

機材エントリーの検索。以下のクエリパラメータを指定して、検索できる。

Query                      | Description
-----                      | -----------
`serial=<serial>`          | 機材のシリアル番号
`datacenter=<datacenter>`  | 機材が置かれているデータセンター
`rack=<rack>`              | 機材が置かれているラック番号
`role=<role>`              | 機材のロール
`index-in-rack=<rack>`     | ラック内の機材を一意に示すインデックス(物理的な場所とは無関係)
`product=<product>`        | 機材の製品名(R630等)
`ipv4=<ip address>`        | IPv4アドレス
`ipv6=<ip address>`        | IPv6アドレス

**成功時のレスポンス**

- ステータスコード: 200 OK
- ボディ: 機材情報の配列

**失敗ケース**

- 機材が1件も見つからなかった: 404 Not Found

```console
$ curl -XGET 'localhost:8888/api/v1/machines?serial=1234abcd'
[{"serial":"1234abcd","product":"R630","datacenter":"us","rack":1,"index-in-rack":1,"role":"boot","network":{"node0":{"ipv4":["10.69.0.197"],"ipv6":null},"node1":{"ipv4":["10.69.1.5"],"ipv6":null},"node2":{"ipv4":["10.69.1.69"],"ipv6":null}},"bmc":{"ipv4":["10.72.17.37"]}}]
$ curl -XGET 'localhost:8888/api/v1/machines?datacenter=ty3&rack=1&product=R630'
[{"serial":"10000000","product":"R630","datacenter":"ty3","rack":1,"index-in-rack":1,"role":"boot","network":{"node0":{"ipv4":["10.69.0.197"],"ipv6":null},"node1":{"ipv4":["10.69.1.5"],"ipv6":null},"node2":{"ipv4":["10.69.1.69"],"ipv6":null}},"bmc":{"ipv4":["10.72.17.37"]}},{"serial":"10000001","product":"R630","datacenter":"ty3","rack":1,"index-in-rack":1,"role":"boot","network":{"node0":{"ipv4":["10.69.0.197"],"ipv6":null},"node1":{"ipv4":["10.69.1.5"],"ipv6":null},"node2":{"ipv4":["10.69.1.69"],"ipv6":null}},"bmc":{"ipv4":["10.72.17.37"]}}]
$ curl -XGET 'localhost:8888/api/v1/machines?ipv4=10.20.30.40'
[{"serial":"20000000","product":"R630","datacenter":"us","rack":1,"index-in-rack":1,"role":"boot","network":{"node0":{"ipv4":["10.69.0.197"],"ipv6":null},"node1":{"ipv4":["10.20.30.40"],"ipv6":null},"node2":{"ipv4":["10.69.1.69"],"ipv6":null}},"bmc":{"ipv4":["10.72.17.37"]}}]
```

### `DELETE /api/v1/machines/<serial>`

機材エントリーの削除。
`<serial>`で指定された機材を削除する。

**成功時のレスポンス**

- ステータスコード: 200 OK

**失敗ケース**

- 指定された`<serial>`の機材がなかった: 404 Not Found

```console
$ curl -i -X DELETE 'localhost:8888/api/v1/machines/1234abcd'
(出力なし)
```

### `GET /api/v1/ignitions/<serial>`

CoreOSのIgnition形式で取得

```console
$ curl -XGET localhost:8888/api/v1/ignitions/1234abcd
```
!!! Caution
    現在未実装

### `PUT /api/v1/crypts/<serial>/<path>`

暗号鍵を追加。リクエストボディは生バイナリの鍵データ。

** 成功時のレスポンス **

- ステータスコード: 201 Created
- レスポンスボディ: `application/json`

    ```json
    {"status": 201, "path": <path>}
    ```

** 失敗ケース **

- etcd に `/<prefix>/crypts/<serial>/<path>` が既に存在する(409 Conflict)

```console
$ echo "バイナリ鍵データ" | curl -i -X PUT -d - \
   'http://localhost:8888/api/v1/crypts/1/aaaaa'
HTTP/1.1 201 Created
Content-Type: application/json
Date: Tue, 10 Apr 2018 09:12:12 GMT
Content-Length: 31

{"status": 201, "path":"aaaaa"}
```

### `GET /api/v1/crypts/<serial>/<path>`

機材の特定のディスクの暗号鍵を取得

**成功時のレスポンス**

- ステータスコード: 200 OK
- レスポンスボディ: `application/octet-stream` 生バイナリの鍵データ

**失敗ケース**

- etcdに `/<prefix>/crypts/<serial>/<path>` が存在しない(404 Not Found)


```console
$ curl -i -X GET \
   'http://localhost:8888/api/v1/crypts/1/aaaaa'
HTTP/1.1 200 OK
Content-Type: application/octet-stream
Date: Tue, 10 Apr 2018 09:15:59 GMT
Content-Length: 64

.....
```

### `DELETE /api/v1/crypts/<serial>`

特定の機材の全てのディスクの暗号鍵の削除。
機材の情報自体は削除しないので、後からシリアル番号から再び登録できる。

**成功時のレスポンス**

- ステータスコードは 200 OK
- ボディは削除に成功した鍵の path の配列

```console
$ curl -i -X DELETE \
   'http://localhost:8888/api/v1/crypts/1'
HTTP/1.1 200 OK
Content-Type: application/json
Date: Tue, 10 Apr 2018 09:19:01 GMT
Content-Length: 18

["abdef", "aaaaa"]
```

## `sabactl`

### Usage

```console
$ sabactl [--server http://localhost:8888]
```

Option     | Default value           | Description
------     | -------------           | -----------
`--server` | `http://localhost:8888` | sabakanのURL

### `sabactl machines create`

新しい機材の追加

```console
$ sabactl machines create -f <machine_informations.json>
```

一括で対象機材を追加するときは、以下のようなJSONを用意する。

```json
[
  { "serial": "<serial1>", "datacenter": "<datacenter1>", "rack": <rack1>, "product": "<product1>", "role": "<role1>" },
  { "serial": "<serial2>", "datacenter": "<datacenter2>", "rack": <rack2>, "product": "<product2>", "role": "<role2>" }
]
```

!!! Caution
    現在未実装

### `sabactl machines get`

マッチする機材の一覧表示

```console
$ sabactl machines get [--serial <serial>] [--state <state>] [--datacenter <datacenter>] [--rack <rack>] [--product <product>] [--ipv4 <ip address>] [--ipv6 <ip address>]
```

!!! Caution
    `--state <state>`は機材ライフサイクルが決まるまで実装しない

!!! Caution
    現在未実装

### `sabactl machines remove`

指定したシリアル番号の機材または条件にマッチする機材のエントリーの削除

```console
$ sabactl machines remove [--state <state>]
```

!!! Note
    removeが必要なのは修理待ちや破棄予定の機材で、それらはstatusで表現する。そのためシリアル番号やラック単位でエントリを削除することはないので、stateのみで指定する。オペレーションミスも防げる。対象の機材を削除したい場合は、statusを変更してから行う。

!!! Caution
    機材ライフサイクルが決まるまで実装しない

## etcd のスキーマ設計

以下のキー/バリューをetcdに作成する。

### `<prefix>/machines/<serial>`


- prefix:   sabakan などの文字列
- serial:   機材のシリアル番号

各機材の情報。データはJSON。

```console
$ etcdctl get /sabakan/machines/1234abcd --print-value-only | jq .
{
  "serial": "1234abcd",
  "product": "R630",
  "datacenter": "ty3",
  "rack": 1,
  "index-in-rack": 1,
  "role": "boot",
  "network": {
    "net0": {
      "ipv4": [
        "10.69.0.69"
      ],
      "ipv6": []
    },
    "net1": {
      "ipv4": [
        "10.69.0.133"
      ],
      "ipv6": []
    }
  },
  "bmc": {
    "ipv4": [
      "10.72.17.37"
    ]
  }
```

Key              | Description
---              | -----------
`serial`         | 機材のシリアル番号
`product`        | 機材の製品名(R630等)
`datacenter`     | 機材が置かれているデータセンター
`rack`           | 機材が置かれているラックの論理ラック番号(LRN)
`index-in-rack`  | ラック内の機材を一意に示すインデックス(物理的な場所とは無関係)
`network`        | NIC名がKeyでIPアドレスがvalueの辞書
`bmc`            | BMC(iDRAC)のIPアドレス

### `<prefix>/crypts/<serial>/<path>`

Name   | Description
----   | -----------
prefix | sabakan などの文字列
serial | 機材のシリアル番号
path   | `/dev/disk/by-path` の下のファイル名

機材の各ディスクの暗号鍵。
データは生バイナリの鍵データ。

```console
$ etcdctl get /sabakan/crypts/1234abcd/pci-0000:00:1f.2-ata-3 --print-value-only
(バイナリ鍵)
```

### `<prefix>/config`

IPAMの設定。

Field                      | Description
------                     | -----------
`max-nodes-in-rack`        | ラック内のWorker Nodeの最大数。Boot Serverは含めない。
`node-ipv4-offset`         | Nodeに割り当てるIPアドレス範囲。
`node-rack-shift`          | `node-ipv4-offset`で指定した範囲を元にNode毎のIPアドレスを算出するための値。
`node-index-offset`        | Nodeに割り当てるインデックスのオフセット。Role が `boot` の Node にはインデックスとして `node-index-offset` を割り当てる。その他の Node には `node-index-offset + 1` 以降の値を割り当てる。 [ネットワーク設計を参照](network_design.md#node-0)
`bmc-ipv4-offset`          | BMC(Baseboard Management Controller)のIPアドレス範囲。NecoではiDRACのIPアドレスに使用する。
`bmc-rack-shift`           | `bmc-ipv4-offset`で指定した範囲を元にBMC毎のIPアドレスを算出するための値。
`node-ip-per-node`         | Node毎に割り当てるIPアドレスの数。BMCは含めない。
`bmc-ip-per-node`          | BMC毎に割り当てるIPアドレスの数。


```console
$ etcdctl get /sabakan/config/ --print-value-only | jq .
{
  "max-nodes-in-rack": 28,
  "node-ipv4-offset": "10.69.0.0/26",
  "node-rack-shift": 6,
  "node-index-offset": 3,
  "node-ip-per-node": 3,
  "bmc-ipv4-offset": "10.72.17.0/27",
  "bmc-rack-shift": 5,
  "bmc-ip-per-node": 1
}
```

### `<prefix>/node-indices/<rack>`

* rack: ラックの番号

ラックごとのノード割り当て情報を登録する。
割り当てたノードインデックスのリストを値とする。

例:
```
$ etcdctl get "/sabakan/node-indices/0"
[3, 4, 5]
```
