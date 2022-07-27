package test

import (
	"github.com/dmznlin/znlib-go/znlib"
	"testing"
)

func TestDes(t *testing.T) {
	var key = []byte("1234567890123456")
	var str = []byte("sa")
	var crypt *znlib.Encrypter = znlib.NewEncrypter(znlib.EncryptAES_ECB, str, key)

	data, err := crypt.Encrypt(true)
	if err != nil || !znlib.Equal(data, []byte("MdRTvC/0/iEjl8K1waWtIw==")) {
		t.Errorf("znlib.AESEncrypt wrong")
	} else {
		t.Log(string(data))
	}

	crypt.Data = data
	data, err = crypt.Decrypt(true)
	if err != nil || !znlib.Equal(data, str) {
		t.Errorf("znlib.AESDecrypt wrong")
	} else {
		t.Log(string(data))
	}

	crypt.Method = znlib.EncryptDES_ECB
	crypt.NewData([]byte("sa"))
	crypt.NewKey([]byte(znlib.DBEncryptKey))

	data, err = crypt.Encrypt(true)
	if err != nil || !znlib.Equal(data, []byte("jKwUUfac8V4=")) {
		t.Errorf("znlib.DESEncrypt wrong")
	} else {
		t.Log(string(data))
	}

	crypt.Data = data
	data, err = crypt.Decrypt(true)
	if err != nil || !znlib.Equal(data, []byte("sa")) {
		t.Errorf("znlib.DESDecrypt wrong")
	} else {
		t.Log(string(data))
	}
}
