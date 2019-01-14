package broker

import (
	"io"
	"log"
	"os"
	"path"
	"time"

	"cloud.google.com/go/storage"
	"github.com/nerikeshi-k/benienma/config"
	"github.com/nerikeshi-k/benienma/utils"
	"golang.org/x/net/context"
)

var ctx context.Context
var googleCloudStorageClient *storage.Client

// 起動時に呼び出される
func init() {
	// ctxはグローバル変数
	ctx = context.Background() // do not change to :=
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	googleCloudStorageClient = client
}

// fetchObject Object NameをキーにRedisのレコードを探したけどなかったとき実行する
// バケットの中から該当nameのオブジェクトをダウンロードしてきてローカルにキャッシュす
// もろもろ終わったら、メタ情報（場所とか名前とか）の構造体を返却する
func fetchObject(name string) (Object, error) {
	var object Object
	var bucketName = config.Get().Bucket.Name

	// ストレージにあるバケット（のハンドラ）
	bkt := googleCloudStorageClient.Bucket(bucketName)

	// バケットからオブジェクトを取ってくる
	obj := bkt.Object(name)
	r, err := obj.NewReader(ctx)
	if err != nil {
		return object, err
	}
	defer r.Close() // defer で終了時に実行

	// オブジェクトのメタ情報
	objAttrs, err := obj.Attrs(ctx)
	if err != nil {
		return object, err
	}

	// キャッシュの保存先を作る
	cachedName := utils.GenerateSafeName() + utils.ExtractExtension(objAttrs.Name)
	fp, err := os.Create(path.Join(config.Get().CacheDirPath, cachedName))
	if err != nil {
		return object, err
	}
	defer fp.Close()

	// 作った保存先に拾ってくるオブジェクトを書き込む
	if _, err := io.Copy(fp, r); err != nil {
		return object, err
	}

	// 場所とかのメタデータ　RedisにValueとして保存される
	object = Object{
		Identity:        objAttrs.Name,
		CachedName:      cachedName,
		Path:            path.Join(config.Get().CacheDirPath, cachedName),
		Size:            objAttrs.Size,
		LastRequestedAt: time.Now(),
		CreatedAt:       time.Now(),
	}

	return object, nil
}
