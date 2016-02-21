package dencrypt

import "testing"

func TestXOREndecode(t *testing.T) {
	data := []byte("hello world")
	mirror := make([]byte, len(data))
	copy(mirror, data)
	if string(mirror) != string(data) {
		t.Fail()
	}

	key := []byte("encode key")
	XOREndecode(data, key)
	t.Log(string(data))
	XOREndecode(data, key)
	t.Log(string(data))
	if string(mirror) != string(data) {
		t.Fail()
	}
}
