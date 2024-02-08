// Package znlib
/******************************************************************************
  作者: dmzn@163.com 2022-07-26 19:53:28
  描述: 数据加解密、数据编码
******************************************************************************/
package znlib

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"github.com/forgoer/openssl"
	"github.com/pkg/errors"
	"io"
)

// EncryptMethod 加密算法
type EncryptMethod = byte

const (
	EncryptAES_ECB EncryptMethod = iota
	EncryptAES_CBC
	//AES: 密钥的长度可以是 16/24/32 个字符（128/192/256 位）

	EncryptDES_ECB
	EncryptDES_CBC
	//DES: 密钥的长度必须为 8 个字符（64 位）

	Encrypt3DES_ECB
	Encrypt3DES_CBC
	//3DES: 密钥的长度必须为 24 个字符（192 位）

	EncryptBase64_STD
	EncryptBase64_URl
	//Base64: StdEncoding,URLEncoding
)

type Encrypter struct {
	Key     []byte        //秘钥
	Padding string        //填充
	Method  EncryptMethod //算法
}

// NewEncrypter 2022-07-27 21:25:31
/*
 参数: method,加密算法
 参数: key,秘钥
 参数: padding,填充模式(PKCS5_PADDING,PKCS7_PADDING,ZEROS_PADDING)
 描述: 生成编码器
*/
func NewEncrypter(method EncryptMethod, key []byte, padding ...string) *Encrypter {
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
		Key:     key,
		Padding: pad,
		Method:  method,
	}
}

// NewKey 2022-07-27 23:22:37
/*
 参数: key,秘钥
 描述: 深层复制秘钥
*/
func (cyp *Encrypter) NewKey(key []byte) {
	cyp.Key = make([]byte, len(key))
	copy(cyp.Key, key)
}

// Encrypt 2022-07-27 21:25:31
/*
 参数: data,数据
 参数: encode,是否base64编码
 参数: iv,加密向量
 描述: 数据加密
*/
func (cyp *Encrypter) Encrypt(data []byte, encode bool, iv ...[]byte) (dst []byte, err error) {
	var iv_data []byte
	switch cyp.Method {
	case EncryptAES_CBC, EncryptDES_CBC, Encrypt3DES_CBC:
		if iv == nil {
			iv_data = cyp.Key
		} else {
			iv_data = iv[0]
		}
	default:
		iv_data = make([]byte, 0)
	}

	switch cyp.Method {
	case EncryptAES_ECB:
		dst, err = openssl.AesECBEncrypt(data, cyp.Key, cyp.Padding)
	case EncryptAES_CBC:
		dst, err = openssl.AesCBCEncrypt(data, cyp.Key, iv_data, cyp.Padding)
	case EncryptDES_ECB:
		dst, err = openssl.DesECBEncrypt(data, cyp.Key, cyp.Padding)
	case EncryptDES_CBC:
		dst, err = openssl.DesCBCEncrypt(data, cyp.Key, iv_data, cyp.Padding)
	case Encrypt3DES_ECB:
		dst, err = openssl.Des3ECBEncrypt(data, cyp.Key, cyp.Padding)
	case Encrypt3DES_CBC:
		dst, err = openssl.Des3CBCEncrypt(data, cyp.Key, iv_data, cyp.Padding)
	default:
		return nil, ErrorMsg(nil, "znlib.encrypt.Encrypt: invalid method.")
	}

	if err == nil && encode {
		return NewEncrypter(EncryptBase64_STD, nil).EncodeBase64(dst)
	}
	return
}

// Decrypt 2022-07-27 21:25:31
/*
 参数: data,数据
 参数: 是否base64编码
 参数: 加密向量
 描述: 数据解密
*/
func (cyp *Encrypter) Decrypt(data []byte, encode bool, iv ...[]byte) (dst []byte, err error) {
	if encode {
		data, err = NewEncrypter(EncryptBase64_STD, nil).DecodeBase64(data)
		if err != nil {
			return nil, err
		}
	}

	var iv_data []byte
	switch cyp.Method {
	case EncryptAES_CBC, EncryptDES_CBC, Encrypt3DES_CBC:
		if iv == nil {
			iv_data = cyp.Key
		} else {
			iv_data = iv[0]
		}
	default:
		iv_data = make([]byte, 0)
	}

	switch cyp.Method {
	case EncryptAES_ECB:
		dst, err = openssl.AesECBDecrypt(data, cyp.Key, cyp.Padding)
	case EncryptAES_CBC:
		dst, err = openssl.AesCBCDecrypt(data, cyp.Key, iv_data, cyp.Padding)
	case EncryptDES_ECB:
		dst, err = openssl.DesECBDecrypt(data, cyp.Key, cyp.Padding)
	case EncryptDES_CBC:
		dst, err = openssl.DesCBCDecrypt(data, cyp.Key, iv_data, cyp.Padding)
	case Encrypt3DES_ECB:
		dst, err = openssl.Des3ECBDecrypt(data, cyp.Key, cyp.Padding)
	case Encrypt3DES_CBC:
		dst, err = openssl.Des3CBCDecrypt(data, cyp.Key, iv_data, cyp.Padding)
	default:
		return nil, ErrorMsg(nil, "znlib.encrypt.Decrypt: invalid method.")
	}

	return
}

// EncodeBase64 2022-07-28 11:51:24
/*
 参数: data,数据
 描述: base64编码
*/
func (cyp *Encrypter) EncodeBase64(data []byte) (dst []byte, err error) {
	var encoding *base64.Encoding
	switch cyp.Method {
	case EncryptBase64_STD:
		encoding = base64.StdEncoding
	case EncryptBase64_URl:
		encoding = base64.URLEncoding
	default:
		return nil, ErrorMsg(nil, "znlib.encrypt.EncodeBase64: invalid method.")
	}

	dst = make([]byte, encoding.EncodedLen(len(data)))
	encoding.Encode(dst, data)
	return dst, nil
}

// DecodeBase64 2022-07-28 11:51:53
/*
 参数: data,数据
 描述: base64解码
*/
func (cyp *Encrypter) DecodeBase64(data []byte) (dst []byte, err error) {
	var encoding *base64.Encoding
	switch cyp.Method {
	case EncryptBase64_STD:
		encoding = base64.StdEncoding
	case EncryptBase64_URl:
		encoding = base64.URLEncoding
	default:
		return nil, ErrorMsg(nil, "znlib.encrypt.DecodeBase64: invalid method.")
	}

	var num int
	dst = make([]byte, encoding.DecodedLen(len(data)))
	num, err = encoding.Decode(dst, data)
	return dst[:num], err
}

//  ---------------------------------------------------------------------------

// ZipUtils zip算法压缩器
type ZipUtils struct{}

// NewZipper 2024-02-08 15:45:42
/*
 描述: 返回支持zip算法的压缩器
*/
func NewZipper() *ZipUtils {
	return &ZipUtils{}
}

// ZipData 2024-02-08 14:38:58
/*
 参数: data,数据
 描述: 使用zip算法压缩数据
*/
func (z *ZipUtils) ZipData(data []byte) (zip []byte, err error) {
	buf := new(bytes.Buffer)
	writer := gzip.NewWriter(buf)
	_, err = writer.Write(data)

	err = ErrorJoin(err, writer.Close())
	if err != nil {
		ErrorCaller(err, "znlib.encrypt.ZipData")
		return nil, err
	}

	return buf.Bytes(), nil
}

// UnzipData 2024-02-08 14:39:14
/*
 参数: data,数据
 描述: 使用zip算法解压缩
*/
func (z *ZipUtils) UnzipData(data []byte) (zip []byte, err error) {
	caller := "znlib.encrypt.UnzipData"
	buf := bytes.NewReader(data)
	reader, err := gzip.NewReader(buf)

	if err != nil {
		ErrorCaller(err, caller)
		return nil, err
	}

	defer func() {
		err = ErrorJoin(err, reader.Close())
		if err != nil {
			ErrorCaller(err, caller)
		}
	}()

	var dt []byte
	dt, err = io.ReadAll(reader)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
		return nil, err
	}

	return dt, nil
}
