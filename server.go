package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo"
	"github.com/nerikeshi-k/benienma/broker"
	"github.com/nerikeshi-k/benienma/config"
	"github.com/nerikeshi-k/benienma/gc"
	"github.com/nerikeshi-k/benienma/image"
)

// メイン関数
func main() {
	e := echo.New()

	serverHeader := func(hf echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set("Last-Modified", config.Get().LastModified)
			c.Response().Header().Set("Access-Control-Allow-Origin", "*")
			return hf(c)
		}
	}
	e.Use(serverHeader)

	// ExpiredなObjectの回収を開始
	go gc.StartCollecting()

	e.GET("/", index)
	e.GET("/order/", manager)
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", config.Get().Port)))
}

func index(c echo.Context) error {
	return c.String(http.StatusOK, "running")
}

func manager(c echo.Context) error {
	// identity 取得　blob名でありredisのキー
	identity := c.QueryParam("name")
	if identity == "" {
		return c.String(http.StatusBadRequest, fmt.Sprintf("invalid"))
	}

	// If-Modified-Since ヘッダの情報を確認して存在してるなら 304
	IfModifiedSince := c.Request().Header.Get("If-Modified-Since")
	if IfModifiedSince == config.Get().LastModified {
		return c.NoContent(http.StatusNotModified)
	}

	// 画像処理の設定を取得
	atoi := func(s string) uint {
		num, err := strconv.ParseUint(s, 10, 32)
		if err != nil {
			return 0
		}
		return uint(num)
	}
	orderDetails := image.OrderDetails{
		MaxWidth:  atoi(c.QueryParam("maxwidth")),
		MaxHeight: atoi(c.QueryParam("maxheight")),
		Width:     atoi(c.QueryParam("width")),
		Height:    atoi(c.QueryParam("height")),
	}

	object, err := broker.Get(identity)

	fmt.Printf("%v", orderDetails)

	if err != nil {
		switch err.(type) {
		case *broker.NotFound:
			return c.String(http.StatusNotFound, fmt.Sprintf("no such object as %s", identity))
		default:
			return c.String(http.StatusInternalServerError, fmt.Sprintf("500: %v", err))
		}
	} else {
		// 画像配信
		return image.ProcessedImageResponse(c, object.Path, orderDetails)
	}
}
