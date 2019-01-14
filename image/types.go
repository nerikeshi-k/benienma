package image

// 画像処理時の設定
type OrderDetails struct {
	MaxWidth  uint // 最大width
	MaxHeight uint // 最大height
	Width     uint // width
	Height    uint // height
}

func (o *OrderDetails) IsDefault() bool {
	if o.MaxWidth == 0 && o.MaxHeight == 0 &&
		o.Width == 0 && o.Height == 0 {
		return true
	} else {
		return false
	}

}
