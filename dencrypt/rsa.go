package dencrypt

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"hash"
	"io"
	"io/ioutil"
)

var (
	NewMd5Hash    = func() hash.Hash { return md5.New() }
	NewSha1Hash   = func() hash.Hash { return sha1.New() }
	NewSha256Hash = func() hash.Hash { return sha256.New() }
)

type RsaEncrypt struct {
	privateKey  *rsa.PrivateKey
	keyBytes    int // for 1024 bits private key it's 128 bytes, used to split and encrypt log message
	maxMsgBytes int
	newHash     func() hash.Hash
}

func NewRsaEncrypt(privateKeyInput io.Reader, keyBytes int, newHash func() hash.Hash) (*RsaEncrypt, error) {
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
	h := newHash()
	r := &RsaEncrypt{
		privateKey:  privateKey,
		keyBytes:    keyBytes,
		maxMsgBytes: keyBytes - (h.Size()*2 + 2),
		newHash:     newHash,
	}
	return r, nil
}

func (r *RsaEncrypt) EncryptOAEP(plainText, label []byte) (encrypted []byte, err error) {
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

		if buf, err = rsa.EncryptOAEP(r.newHash(), rand.Reader, publicKey, plainText[start:start+n], label); err != nil {
			return nil, err
		}
		encrypted = append(encrypted, buf...)
		start += n
	}
	return
}

func (r *RsaEncrypt) DecryptOAEP(encrypted, label []byte) (decrypted []byte, err error) {
	var buf []byte
	sz := len(encrypted)
	start := 0
	for start < sz {
		n := r.keyBytes
		left := sz - start
		if n > left {
			n = left
		}

		if buf, err = rsa.DecryptOAEP(r.newHash(), rand.Reader, r.privateKey, encrypted[start:start+n], label); err != nil {
			return nil, err
		}
		decrypted = append(decrypted, buf...)
		start += n
	}
	return
}

type RsaPKCSEncrypt struct {
	privateKey  *rsa.PrivateKey
	keyBytes    int // for 1024 bits private key it's 128 bytes, used to split and encrypt log message
	maxMsgBytes int
}

func NewRsaPKCSEncrypt(privateKeyInput io.Reader, keyBytes int) (*RsaPKCSEncrypt, error) {
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
	r := &RsaPKCSEncrypt{
		privateKey:  privateKey,
		keyBytes:    keyBytes,
		maxMsgBytes: keyBytes - 11,
	}
	return r, nil
}

func (r *RsaPKCSEncrypt) EncryptPKCS1v15(plainText []byte) (encrypted []byte, err error) {
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

		if buf, err = rsa.EncryptPKCS1v15(rand.Reader, publicKey, plainText[start:start+n]); err != nil {
			return nil, err
		}
		encrypted = append(encrypted, buf...)
		start += n
	}
	return
}

func (r *RsaPKCSEncrypt) DecryptPKCS1v15(encrypted []byte) (decrypted []byte, err error) {
	var buf []byte
	sz := len(encrypted)
	start := 0
	for start < sz {
		n := r.keyBytes
		left := sz - start
		if n > left {
			n = left
		}

		if buf, err = rsa.DecryptPKCS1v15(rand.Reader, r.privateKey, encrypted[start:start+n]); err != nil {
			return nil, err
		}
		decrypted = append(decrypted, buf...)
		start += n
	}
	return
}
