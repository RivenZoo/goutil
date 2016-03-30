package dencrypt

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"testing"
)

func testRsaEncrypt() *RsaEncrypt {
	var key *rsa.PrivateKey
	var err error

	keyBytes := 128
	if key, err = rsa.GenerateKey(rand.Reader, keyBytes*8); err != nil {
		return nil
	}
	rsaEncrypt := &RsaEncrypt{
		privateKey:  key,
		keyBytes:    keyBytes,
		maxMsgBytes: keyBytes - (defaultHashBytes*2 + 2),
	}
	return rsaEncrypt
}

func TestRsaAndBase64(t *testing.T) {
	plainText := []byte(`{"num":"100001","tunnel":"13034","data":{"id":1,"md5":"f5148ac391c2bfbcc6dd6a5bb754612c"}}`)
	rsaEncrypt := testRsaEncrypt()
	encrypted, err := rsaEncrypt.EncryptOAEP(plainText, nil)
	if err != nil {
		t.Fatal(err)
	}
	s := base64.StdEncoding.EncodeToString(encrypted)
	t.Log(s)
}
func TestRsaEnDecrypt(t *testing.T) {
	plainText := make([]byte, 1024)
	rand.Read(plainText)
	rsaEncrypt := testRsaEncrypt()
	encrypted, err := rsaEncrypt.EncryptOAEP(plainText, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(encrypted, len(encrypted))
	decrypted, err := rsaEncrypt.DecryptOAEP(encrypted, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(decrypted)
	if !bytes.Equal(decrypted, plainText) {
		t.Fail()
	}
}

func BenchmarkRsaEncrypt(b *testing.B) {
	plainText := make([]byte, 1024)
	rand.Read(plainText)
	rsaEncrypt := testRsaEncrypt()
	for i := 0; i < b.N; i++ {
		_, err := rsaEncrypt.EncryptOAEP(plainText, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRsaEncryptParallel(b *testing.B) {
	plainText := make([]byte, 1024)
	rand.Read(plainText)
	rsaEncrypt := testRsaEncrypt()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := rsaEncrypt.EncryptOAEP(plainText, nil)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkRsaDecrypt(b *testing.B) {
	plainText := make([]byte, 1024)
	rand.Read(plainText)
	rsaEncrypt := testRsaEncrypt()
	encrypted, err := rsaEncrypt.EncryptOAEP(plainText, nil)
	if err != nil {
		b.Fatal(err)
	}
	for i := 0; i < b.N; i++ {
		_, err := rsaEncrypt.DecryptOAEP(encrypted, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRsaDecryptParallel(b *testing.B) {
	plainText := make([]byte, 1024)
	rand.Read(plainText)
	rsaEncrypt := testRsaEncrypt()
	encrypted, err := rsaEncrypt.EncryptOAEP(plainText, nil)
	if err != nil {
		b.Fatal(err)
	}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := rsaEncrypt.DecryptOAEP(encrypted, nil)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
