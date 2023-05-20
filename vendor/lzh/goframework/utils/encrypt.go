package utils

import (
	"bytes"
	"crypto/des"
	"crypto/md5"
	"crypto/rc4"
	"encoding/hex"
	"errors"
	"fmt"
)

/*
	单向加密算法
*/
func Md5(plain string) string {
	h := md5.New()
	h.Write([]byte(plain))
	cipherStr := h.Sum(nil)
	return hex.EncodeToString(cipherStr)
}

func SHA1() {

}

func SHA256() {

}

/*
	对称加密算法
*/

const (
	PaddingTypePKCS7 = iota // 推荐使用
	PaddingTypePKCS5
	PaddingTypeZero      // 不推荐使用
	PaddingTypeNoPadding // 不要用
)

// ECB加密, 使用PKCS7进行填充
func EncryptDesEcb(src, key string, paddingType int) (e string, err error) {
	data := []byte(src)
	keyByte := []byte(key)
	block, err := des.NewCipher(keyByte)
	if err != nil {
		return e, err
	}
	bs := block.BlockSize()
	// 对明文数据进行补码
	data = padding(paddingType, data, bs)
	if len(data)%bs != 0 {
		return e, errors.New("need a multiple of the block size")
	}
	out := make([]byte, len(data))
	dst := out
	for len(data) > 0 {
		// 对明文按照 blockSize 进行分块加密
		block.Encrypt(dst, data[:bs])
		data = data[bs:]
		dst = dst[bs:]
	}
	e = fmt.Sprintf("%x", out)
	return e, nil
}

// ECB解密, 使用PKCS7进行填充
func DecryptDesEcb(src, key string, paddingType int) (d string, err error) {
	data, err := hex.DecodeString(src)
	if err != nil {
		return d, errors.New(fmt.Sprintf("hex.DecodeString error: %s", err.Error()))
	}
	keyByte := []byte(key)
	block, err := des.NewCipher(keyByte)
	if err != nil {
		return d, errors.New(fmt.Sprintf("des.NewCipher error: %s", err.Error()))
	}
	bs := block.BlockSize()
	if len(data)%bs != 0 {
		return d, errors.New("crypto/cipher: input not full blocks")
	}
	out := make([]byte, len(data))
	dst := out
	for len(data) > 0 {
		block.Decrypt(dst, data[:bs])
		data = data[bs:]
		dst = dst[bs:]
	}
	out = unPacking(paddingType, out)
	d = string(out)
	return d, nil
}

func padding(packType int, cipherText []byte, blockSize int) (out []byte) {
	switch packType {
	case PaddingTypePKCS7:
		out = PKCS7Padding(cipherText, blockSize)
	case PaddingTypePKCS5:
		out = PKCS5Padding(cipherText)
	case PaddingTypeZero:
		out = ZeroPadding(cipherText, blockSize)
	}

	return out
}

func unPacking(packType int, origData []byte) (out []byte) {
	switch packType {
	case PaddingTypePKCS7:
		out = PKCS7UnPadding(origData)
	case PaddingTypePKCS5:
		out = PKCS5UnPadding(origData)
	//case PaddingTypeZero:
	// 兼容php使用，不用unPad
	//out = ZeroUnPadding(origData)
	default:
		out = origData
	}
	return out
}

// 假设数据长度需要填充n(n>0)个字节才对齐，那么填充n个字节，每个字节都是n;
// 如果数据本身就已经对齐了，则填充一块长度为块大小的数据，每个字节都是块大小
func PKCS7Padding(cipherText []byte, blockSize int) []byte {
	padding := blockSize - len(cipherText)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(cipherText, padText...)
}

func PKCS7UnPadding(origData []byte) []byte {
	size := len(origData)
	if size == 0 {
		return nil
	}

	padding := int(origData[size-1])
	if padding < 1 || padding > 32 {
		padding = 0
	}
	return origData[:size-padding]
}

// PKCS7Padding的子集，块大小固定为8字节
func PKCS5Padding(cipherText []byte) []byte {
	return PKCS7Padding(cipherText, 8)
}

func PKCS5UnPadding(origData []byte) []byte {
	return PKCS7UnPadding(origData)
}

// ZeroPadding，数据长度不对齐时使用0填充，否则不填充
// 只适合以\0结尾的字符串加解密
func ZeroPadding(cipherText []byte, blockSize int) []byte {
	padding := blockSize - len(cipherText)%blockSize
	padText := bytes.Repeat([]byte{byte(0)}, padding)
	return append(cipherText, padText...)
}
func ZeroUnPadding(origData []byte) []byte {
	return bytes.TrimRightFunc(origData,
		func(r rune) bool {
			return r == rune(0)
		})
}
func Rc4EncodeBytes(key, plainText []byte) (string, error) {
	c, err := rc4.NewCipher(key)
	if err != nil {
		return "", err
	}

	dst := make([]byte, len(plainText))
	c.XORKeyStream(dst, plainText)
	return hex.EncodeToString(dst), nil
}

func Rc4DecodeBytes(key []byte, encrypted string) ([]byte, error) {
	src, err := hex.DecodeString(encrypted)
	if err != nil {
		return nil, err
	}

	plain := make([]byte, len(src))
	cipher2, err := rc4.NewCipher(key)
	if err != nil {
		return nil, err
	}
	cipher2.XORKeyStream(plain, src)
	return plain, nil
}

func Rc4Encode(key, plainText string) (string, error) {
	src := StringToBytes(plainText)
	k := StringToBytes(key)

	return Rc4EncodeBytes(k, src)
}

func Rc4Decode(key, encrypted string) (string, error) {
	k := StringToBytes(key)
	plain, err := Rc4DecodeBytes(k, encrypted)
	if err != nil {
		return "", err
	}
	return BytesToString(plain), nil
}
