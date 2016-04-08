package dencrypt

import (
	"bytes"
	"testing"
	"encoding/base64"
)

func TestCFBEncrypt(t *testing.T) {
	// aesKey := GetAesKey(0)[:]
	aesKey := []byte("GiH5(yLCVaZR4^-A")
	plain := []byte("{}")
	iv := aesKey

	CFBEncrypt(aesKey, iv, plain)
	t.Log(plain, string(aesKey))
	CFBDecrypt(aesKey, iv, plain)
	if !bytes.Equal(plain, []byte("{}")) {
		t.Error("CFBEncrypt test failed.")
	}
}

func BenchmarkGenRandKey(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenRandKey(uint32(i))
	}
}

func BenchmarkGenTimeKey(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenTimeKey()
	}
}

func TestAESECB(t *testing.T) {
	data := "abbbbbbbbbbbbbbbccbbb"
	key := "1111111111111111"
	expect := "MKkg0sg2vrFLcreSGyJ7nVPXRpt3TGoWFPl+ykXruO8="
	buf, err := ECBEncrypt([]byte(key), []byte(data))
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	ret := base64.StdEncoding.EncodeToString(buf)
	if ret != expect {
		t.Log(ret)
		t.FailNow()
	}
	plain, err := ECBDecrypt([]byte(key), buf)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	if string(plain) != data {
		t.Log(string(plain))
		t.FailNow()
	}
}