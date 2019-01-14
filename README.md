# benienma
GoogleCloudStorageにある画像を配信するサーバ  
キャッシュで通信量を節約する用（cloudflareでよくない？）

# 起動例
```
$ go get github.com/nerikeshi-k/benienma
$ cd $GOPATH/src/github.com/nerikeshi-k/benienma
$ go build
$ cp config.default.json config.json
$ vi config.json # よしなに
$ GOOGLE_APPLICATION_CREDENTIALS=/Path/to/gcp/key.json ./benienma --conf=./config.json
```
