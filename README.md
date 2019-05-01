# benienma
Google Cloud Storageにある画像を配信する、いわゆる画像配信サーバです。  
直配信を避けることで通信量を節約できる……かも。cloudflareで十分かもしれない。  
キャッシュと配信時リサイズに対応しています。それ以上の機能は未定です。

## スペック概略

- Google Cloud Storage専用の簡単な画像配信サーバーアプリケーションです
- 複数台立てればCDNのように使えます（未実践）
- 対応画像形式はPNG, JPEGのみです
- キャッシュの能動的なパージはできません
- 複数bucketにまたがって配信させることはできません
- 対応はGoogle Cloud Storageのみで、AWSやAzureでは使えません

## 起動例
事前にやること

- go自体のインストール
- Redisのインストールと設定
- キャッシュ画像を保存するディレクトリの作成

それから↓

```
$ git clone https://github.com/nerikeshi-k/benienma.git
$ cd benienma
$ go build
$ cp config.default.json config.json
$ vi config.json # よしなに
$ GOOGLE_APPLICATION_CREDENTIALS=/Path/to/gcp/key.json ./benienma --conf=./config.json
```

## URLの叩き方
以下のようにURLを叩くことでGoogle Cloud Storageに存在する画像を配信することができます。

```
/order/?name=${オブジェクトのname}
# e.g.
# /order/?name=photos/userA/hogehoge.jpeg
```

一度配信した画像は、一定期間過ぎるかキャッシュ用ストレージがいっぱいになるまでの間キャッシュします。  

### リサイズ
GETパラメータを次のように指定します。

```
/order/?name=${オブジェクトのURL}&maxwidth={数値}&maxheight={数値}
# e.g.
# /order/?name=photos/userA/hogehoge.jpeg&maxwidth=100&maxheight=100
```

リサイズ処理はURLが叩かれるたびに行なわれてしまうので、前段にクエリ込みでキャッシュするように設定した
cloudflareなどを挟んだほうがいいかもしれません。
