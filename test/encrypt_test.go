package test

import (
	"fmt"
	"github.com/dmznlin/znlib-go/znlib"
	"testing"
)

func TestDes(t *testing.T) {
	var str = []byte("sa")
	var key = []byte("1234567890123456")
	var crypt *znlib.Encrypter = znlib.NewEncrypter(znlib.EncryptAES_ECB, key)

	data, err := crypt.Encrypt(str, true)
	if err != nil || !znlib.Equal(data, []byte("MdRTvC/0/iEjl8K1waWtIw==")) {
		t.Errorf("znlib.AESEncrypt wrong")
	} else {
		t.Log(string(data))
	}

	data, err = crypt.Decrypt(data, true)
	if err != nil || !znlib.Equal(data, str) {
		t.Errorf("znlib.AESDecrypt wrong")
	} else {
		t.Log(string(data))
	}
	//------------------------------------------------------------------------------------------------------------------
	crypt.Method = znlib.EncryptDES_ECB
	crypt.NewKey([]byte(znlib.DefaultEncryptKey))

	data, err = crypt.Encrypt(str, true)
	if err != nil || !znlib.Equal(data, []byte("sMGRVV9wABI=")) {
		t.Errorf("znlib.DESEncrypt wrong")
	} else {
		t.Log(string(data))
	}

	data, err = crypt.Decrypt(data, true)
	if err != nil || !znlib.Equal(data, []byte("sa")) {
		t.Errorf("znlib.DESDecrypt wrong")
	} else {
		t.Log(string(data))
	}
	//------------------------------------------------------------------------------------------------------------------
	crypt.Method = znlib.EncryptBase64_STD
	data, err = crypt.EncodeBase64(str)
	if err != nil || !znlib.Equal(data, []byte("c2E=")) {
		t.Errorf("znlib.EncodeBase64 wrong")
	} else {
		t.Log(string(data))
	}

	data, err = crypt.DecodeBase64(data)
	if err != nil || !znlib.Equal(data, []byte("sa")) {
		t.Errorf("znlib.DecodeBase64 wrong")
	} else {
		t.Log(string(data))
	}

	//------------------------------------------------------------------------------------------------------------------
	data, err = znlib.NewZipper().ZipData(key)
	if err != nil {
		t.Errorf("znlib.Zipper.ZipData wrong")
	}

	znlib.Info(fmt.Sprintf("%d -> %d", len(key), len(data)))

	var buf []byte
	buf, err = znlib.NewEncrypter(znlib.EncryptBase64_STD, nil).EncodeBase64(data)
	if !znlib.Equal(buf, []byte("H4sIAAAAAAAA/zI0MjYxNTO3sDSAsAABAAD//7fNXx4QAAAA")) {
		t.Errorf("znlib.Zipper.ZipData wrong")
	}

	buf, err = znlib.NewZipper().UnzipData(data)
	if err != nil || !znlib.Equal(buf, key) {
		t.Errorf("znlib.Zipper.UnzipData wrong")
	}

}
