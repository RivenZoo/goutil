package qrcode

// qr code wrapper

import (
	"image/png"
	"os"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
	"github.com/qpliu/qrencode-go/qrencode"
	"io"
)

type EncodeMode int

const   (
	Auto         = iota
	Numeric
	AlphaNumeric
	Unicode
)

// error correction level
type ECLevel int

const (
	ECLevel_L = iota
	ECLevel_M
	ECLevel_Q
	ECLevel_H
)

type QREncoder struct {
	ecLevel    qr.ErrorCorrectionLevel
	encodeMode qr.Encoding
	size       int
}

func convertECLevel(eclevel ECLevel) qr.ErrorCorrectionLevel {
	lvl := qr.L
	switch eclevel {
	case ECLevel_M:
		lvl = qr.M
	case ECLevel_Q:
		lvl = qr.Q
	case ECLevel_H:
		lvl = qr.H
	}
	return lvl
}

func convertEncodeMode(mode EncodeMode) qr.Encoding {
	encodeMode := qr.Unicode
	switch mode {
	case Auto:
		encodeMode = qr.Auto
	case Numeric:
		encodeMode = qr.Numeric
	case AlphaNumeric:
		encodeMode = qr.AlphaNumeric
	}
	return encodeMode
}

func NewQREncoder(mode EncodeMode, eclevel ECLevel, size int) *QREncoder {
	lvl := convertECLevel(eclevel)
	encodeMode := convertEncodeMode(mode)
	return &QREncoder{
		lvl,
		encodeMode,
		size,
	}
}

func (en *QREncoder) Encode(content string, w io.Writer) (err error) {
	code, err := qr.Encode(content, en.ecLevel, en.encodeMode)
	if err != nil {
		return
	}
	code, err = barcode.Scale(code, en.size, en.size)
	if err != nil {
		return
	}
	err = png.Encode(w, code)
	return
}

func (en *QREncoder) EncodeToFile(content, output string) (err error) {
	f, err := os.Create(output)
	if err != nil {
		return
	}
	defer f.Close()

	code, err := qr.Encode(content, en.ecLevel, en.encodeMode)
	if err != nil {
		return
	}
	code, err = barcode.Scale(code, en.size, en.size)
	if err != nil {
		return
	}
	err = png.Encode(f, code)
	return
}

type alternateQREncoder struct {
	ecLevel   qrencode.ECLevel
	blockSize int
}

func newQREncoder2(eclevel ECLevel, blockSize int) *alternateQREncoder {
	lvl := qrencode.ECLevelL
	switch eclevel {
	case ECLevel_L:
		lvl = qrencode.ECLevelL
	case ECLevel_H:
		lvl = qrencode.ECLevelH
	case ECLevel_M:
		lvl = qrencode.ECLevelM
	case ECLevel_Q:
		lvl = qrencode.ECLevelQ
	}
	return &alternateQREncoder{
		lvl,
		blockSize,
	}
}

func (en *alternateQREncoder) EncodeToFile2(content, output string) (err error) {
	grid, err := qrencode.Encode(content, en.ecLevel)
	if err != nil {
		return
	}
	f, err := os.Create(output)
	if err != nil {
		return
	}
	defer f.Close()
	err = png.Encode(f, grid.Image(en.blockSize))
	return
}

func (en *alternateQREncoder) Encode2(content string, w io.Writer) (err error) {
	grid, err := qrencode.Encode(content, en.ecLevel)
	if err != nil {
		return
	}
	err = png.Encode(w, grid.Image(en.blockSize))
	return
}