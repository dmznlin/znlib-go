/*Package znlib ***************************************************************
  作者: dmzn@163.com 2022-07-26 19:53:28
  描述: 数据加解密、数据编码
******************************************************************************/
package znlib

import (
	"encoding/base64"
	"errors"
	"github.com/forgoer/openssl"
)

//EncryptFlag 加密标识
type EncryptFlag = byte

const (
	EncryptAES_ECB EncryptFlag = iota
	EncryptAES_CBC
	//AES: 密钥的长度可以是 16/24/32 个字符（128/192/256 位）

	EncryptDES_ECB
	EncryptDES_CBC
	//DES: 密钥的长度必须为 8 个字符（64 位）

	Encrypt3DES_ECB
	Encrypt3DES_CBC
	//3DES: 密钥的长度必须为 24 个字符（192 位）
)

type Encrypter struct {
	Data    []byte      //数据
	Key     []byte      //秘钥
	Padding string      //填充
	Method  EncryptFlag //算法
}

/*NewEncrypter 2022-07-27 21:25:31
  参数: method,加密算法
  参数: data,数据
  参数: key,秘钥
  参数: padding,填充模式(PKCS5_PADDING,PKCS7_PADDING,ZEROS_PADDING)
  描述: 生成编码器
*/
func NewEncrypter(method EncryptFlag, data, key []byte, padding ...string) *Encrypter {
	var pad string
	if padding == nil {
		pad = openssl.PKCS7_PADDING
	} else {
		pad = padding[0]
		if !StrIn(pad, openssl.PKCS5_PADDING, openssl.PKCS7_PADDING, openssl.ZEROS_PADDING) {
			pad = openssl.PKCS7_PADDING
		}
	}

	return &Encrypter{
		Data:    data,
		Key:     key,
		Padding: pad,
		Method:  method,
	}
}

/*NewData 2022-07-27 23:21:36
  参数: data,数据
  描述: 深层复制数据
*/
func (cyp *Encrypter) NewData(data []byte) {
	cyp.Data = make([]byte, len(data))
	copy(cyp.Data, data)
}

/*NewKey 2022-07-27 23:22:37
  参数: key,秘钥
  描述: 深层复制秘钥
*/
func (cyp *Encrypter) NewKey(key []byte) {
	cyp.Key = make([]byte, len(key))
	copy(cyp.Key, key)
}

/*Encrypt 2022-07-27 21:25:31
  参数: 是否base64编码
  参数: 加密向量
  描述: 数据加密
*/
func (cyp *Encrypter) Encrypt(encode bool, iv ...[]byte) (dst []byte, err error) {
	var iv_data []byte
	switch cyp.Method {
	case EncryptAES_CBC, EncryptDES_CBC, Encrypt3DES_CBC:
		if iv == nil {
			iv_data = cyp.Key
		} else {
			iv_data = iv[0]
		}
	}
	switch cyp.Method {
	case EncryptAES_ECB:
		dst, err = openssl.AesECBEncrypt(cyp.Data, cyp.Key, cyp.Padding)
	case EncryptAES_CBC:
		dst, err = openssl.AesCBCEncrypt(cyp.Data, cyp.Key, iv_data, cyp.Padding)
	case EncryptDES_ECB:
		dst, err = openssl.DesECBEncrypt(cyp.Data, cyp.Key, cyp.Padding)
	case EncryptDES_CBC:
		dst, err = openssl.DesCBCEncrypt(cyp.Data, cyp.Key, iv_data, cyp.Padding)
	case Encrypt3DES_ECB:
		dst, err = openssl.Des3ECBEncrypt(cyp.Data, cyp.Key, cyp.Padding)
	case Encrypt3DES_CBC:
		dst, err = openssl.Des3CBCEncrypt(cyp.Data, cyp.Key, iv_data, cyp.Padding)
	default:
		return nil, errors.New("znlib.Encrypt: invalid method.")
	}

	if err == nil && encode {
		dst = []byte(base64.StdEncoding.EncodeToString(dst))
	}
	return
}

/*Decrypt 2022-07-27 21:25:31
  参数: 是否base64编码
  参数: 加密向量
  描述: 数据解密
*/
func (cyp *Encrypter) Decrypt(encode bool, iv ...[]byte) (dst []byte, err error) {
	if encode {
		dst = make([]byte, base64.StdEncoding.DecodedLen(len(cyp.Data)))
		var num int
		num, err = base64.StdEncoding.Decode(dst, cyp.Data)

		if err != nil {
			return nil, err
		}

		cyp.Data = dst[:num]
	}

	var iv_data []byte
	switch cyp.Method {
	case EncryptAES_CBC, EncryptDES_CBC, Encrypt3DES_CBC:
		if iv == nil {
			iv_data = cyp.Key
		} else {
			iv_data = iv[0]
		}
	}
	switch cyp.Method {
	case EncryptAES_ECB:
		dst, err = openssl.AesECBDecrypt(cyp.Data, cyp.Key, cyp.Padding)
	case EncryptAES_CBC:
		dst, err = openssl.AesCBCDecrypt(cyp.Data, cyp.Key, iv_data, cyp.Padding)
	case EncryptDES_ECB:
		dst, err = openssl.DesECBDecrypt(cyp.Data, cyp.Key, cyp.Padding)
	case EncryptDES_CBC:
		dst, err = openssl.DesCBCDecrypt(cyp.Data, cyp.Key, iv_data, cyp.Padding)
	case Encrypt3DES_ECB:
		dst, err = openssl.Des3ECBDecrypt(cyp.Data, cyp.Key, cyp.Padding)
	case Encrypt3DES_CBC:
		dst, err = openssl.Des3CBCDecrypt(cyp.Data, cyp.Key, iv_data, cyp.Padding)
	default:
		return nil, errors.New("znlib.Decrypt: invalid method.")
	}

	return
}
