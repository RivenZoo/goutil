package dencrypt

import (
	"bytes"
	"testing"
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
