package dencrypt

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io"
	"io/ioutil"
)

var (
	defaultHashBytes = 16
)

type RsaEncrypt struct {
	privateKey  *rsa.PrivateKey
	keyBytes    int // for 1024 bits private key it's 128 bytes, used to split and encrypt log message
	maxMsgBytes int
}

func NewRsaEncrypt(privateKeyInput io.Reader, keyBytes int) (*RsaEncrypt, error) {
	data, err := ioutil.ReadAll(privateKeyInput)
	if err != nil {
		return nil, err
	}

	var block *pem.Block
	if block, _ = pem.Decode(data); block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, errors.New("wrong private key")
	}

	var privateKey *rsa.PrivateKey
	if privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes); err != nil {
		return nil, err
	}

	privateKey.Precompute()
	if err = privateKey.Validate(); err != nil {
		return nil, err
	}
	r := &RsaEncrypt{
		privateKey:  privateKey,
		keyBytes:    keyBytes,
		maxMsgBytes: keyBytes - (defaultHashBytes*2 + 2),
	}
	return r, nil
}

func (r *RsaEncrypt) EncryptOAEP(plainText, label []byte) (encrypted []byte, err error) {
	md5Hash := md5.New()
	var buf []byte
	publicKey := &r.privateKey.PublicKey

	sz := len(plainText)
	start := 0
	for start < sz {
		n := r.maxMsgBytes
		left := sz - start
		if n > left {
			n = left
		}

		if buf, err = rsa.EncryptOAEP(md5Hash, rand.Reader, publicKey, plainText[start:start+n], label); err != nil {
			return nil, err
		}
		encrypted = append(encrypted, buf...)
		start += n
	}
	return
}

func (r *RsaEncrypt) DecryptOAEP(encrypted, label []byte) (decrypted []byte, err error) {
	md5Hash := md5.New()
	var buf []byte
	sz := len(encrypted)
	start := 0
	for start < sz {
		n := r.keyBytes
		left := sz - start
		if n > left {
			n = left
		}

		if buf, err = rsa.DecryptOAEP(md5Hash, rand.Reader, r.privateKey, encrypted[start:start+n], label); err != nil {
			return nil, err
		}
		decrypted = append(decrypted, buf...)
		start += n
	}
	return
}
