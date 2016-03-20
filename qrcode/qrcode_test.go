package qrcode

import (
	"crypto/md5"
	"encoding/hex"
	"os"
	"path"
	"testing"
	"io/ioutil"
)

const (
	store = "./tmp"
	str   = "http://www.baike.com/wiki/QR%E7%A0%81"
)

func init() {
	os.Mkdir(store, 0777)
}

func TestQREncode(t *testing.T) {
	encoder := NewQREncoder(Unicode, ECLevel_L, 200)
	abstract := md5.Sum([]byte(str))
	fname := hex.EncodeToString(abstract[:]) + ".png"
	err := encoder.EncodeToFile(str, path.Join(store, fname))
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
}

func TestQREncode2(t *testing.T) {
	encoder := newQREncoder2(ECLevel_L, 5)
	abstract := md5.Sum([]byte(str))
	fname := hex.EncodeToString(abstract[:]) + "-2.png"
	err := encoder.EncodeToFile2(str, path.Join(store, fname))
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
}

func BenchmarkQREncode(b *testing.B) {
	encoder := NewQREncoder(Unicode, ECLevel_H, 200)
	var err error
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			err = encoder.Encode(str, ioutil.Discard)
			if err != nil {
				b.FailNow()
			}
		}
	})
}

func BenchmarkQREncode2(b *testing.B) {
	encoder := newQREncoder2(ECLevel_H, 5)
	var err error
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			err = encoder.Encode2(str, ioutil.Discard)
			if err != nil {
				b.FailNow()
			}
		}
	})
}