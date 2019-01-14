package image

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/labstack/echo"
	"github.com/nerikeshi-k/benienma/utils"

	"bytes"
	"image"
	"image/jpeg"
	"image/png"

	"github.com/nfnt/resize"
)

func ProcessedImageResponse(c echo.Context, filepath string, od OrderDetails) error {
	file, err := os.Open(filepath)
	if err != nil {
		return c.String(http.StatusNotFound, fmt.Sprintf("no cached file"))
	}
	defer file.Close()

	frmt := strings.ToLower(utils.ExtractExtension(filepath)[1:]) // eliminate '.'
	// ↑ "jpeg" or "png"

	content_type := fmt.Sprintf("image/%s", frmt)

	// もしOrderDetailsに設定がないならそのまま開いて返す
	// formatについては拡張子を信用する
	if od.IsDefault() {
		byt, err := ioutil.ReadAll(file)
		if err != nil {
			return c.String(http.StatusInternalServerError, fmt.Sprintf("500: %v", err))
		}
		return c.Blob(http.StatusOK, content_type, byt)

	} else {
		// OrderDetailsの設定に準じてリサイズなどをして返す
		var img image.Image
		var err error

		if frmt == "jpeg" {
			img, err = jpeg.Decode(file)
		} else if frmt == "png" {
			img, err = png.Decode(file)
		} else {
			return c.String(http.StatusInternalServerError, "Unknown File Format")
		}

		if err != nil {
			return c.String(http.StatusInternalServerError, fmt.Sprintf("500: %v", err))
		}

		// リサイズ処理
		var m image.Image
		if od.MaxWidth != 0 && od.MaxHeight != 0 {
			m = resize.Thumbnail(od.MaxWidth, od.MaxHeight, img, resize.Lanczos3)

		} else if od.MaxWidth != 0 && od.MaxHeight == 0 {
			m = resize.Thumbnail(od.MaxWidth, 4294967295, img, resize.Lanczos3)

		} else if od.MaxWidth == 0 && od.MaxHeight != 0 {
			m = resize.Thumbnail(4294967295, od.MaxHeight, img, resize.Lanczos3)

		} else if od.Width != 0 {
			m = resize.Resize(od.Width, 0, img, resize.Lanczos3)

		} else if od.Height != 0 {
			m = resize.Resize(0, od.Height, img, resize.Lanczos3)

		}

		buf := new(bytes.Buffer)

		switch frmt {
		case "jpeg":
			err = jpeg.Encode(buf, m, &jpeg.Options{Quality: 100})
		case "png":
			err = png.Encode(buf, m)
		}

		if err != nil {
			return c.String(http.StatusInternalServerError, fmt.Sprintf("500: %v", err))
		}

		return c.Blob(http.StatusOK, content_type, buf.Bytes())
	}
}
