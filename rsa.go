package kuu

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

// GenRSAKey 生成RSA密钥
func GenRSAKey() (prvKey, pubKey []byte) {
	// 生成私钥文件
	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	derStream := x509.MarshalPKCS1PrivateKey(privateKey)
	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: derStream,
	}
	prvKey = pem.EncodeToMemory(block)
	publicKey := &privateKey.PublicKey
	derPkix, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		panic(err)
	}
	block = &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: derPkix,
	}
	pubKey = pem.EncodeToMemory(block)
	return
}

// RSASignWithSha256 签名
func RSASignWithSha256(data []byte, keyBytes []byte) ([]byte, error) {
	h := sha256.New()
	h.Write(data)
	hashed := h.Sum(nil)
	block, _ := pem.Decode(keyBytes)
	if block == nil {
		return nil, errors.New("private key error")
	}
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hashed)
}

// RSAVerySignWithSha256 验签
func RSAVerySignWithSha256(data, signData, keyBytes []byte) error {
	block, _ := pem.Decode(keyBytes)
	if block == nil {
		return errors.New("public key error")
	}
	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return err
	}

	hashed := sha256.Sum256(data)
	err = rsa.VerifyPKCS1v15(pubKey.(*rsa.PublicKey), crypto.SHA256, hashed[:], signData)
	if err != nil {
		return err
	}
	return nil
}

// RSAEncrypt 公钥加密（支持不限长度的明文）
func RSAEncrypt(data, keyBytes []byte) ([]byte, error) {
	//解密pem格式的公钥
	block, _ := pem.Decode(keyBytes)
	if block == nil {
		panic(errors.New("public key error"))
	}
	// 解析公钥
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	// 类型断言
	pub := pubInterface.(*rsa.PublicKey)
	//加密
	partLen := pub.N.BitLen()/8 - 11
	chunks := splitRSAChunks(data, partLen)
	buffer := bytes.NewBufferString("")
	for _, chunk := range chunks {
		data, err := rsa.EncryptPKCS1v15(rand.Reader, pub, chunk)
		if err != nil {
			return nil, err
		}
		buffer.Write(data)
	}
	return buffer.Bytes(), nil
}

// RSADecrypt 私钥解密（此处传入公钥的作用是为了解析不限长度的明文）
func RSADecrypt(ciphertext, pubKeyBytes, privKeyBytes []byte) ([]byte, error) {
	//获取私钥
	privBlock, _ := pem.Decode(privKeyBytes)
	if privBlock == nil {
		return nil, errors.New("private key error")
	}
	//解析PKCS1格式的私钥
	priv, err := x509.ParsePKCS1PrivateKey(privBlock.Bytes)
	if err != nil {
		return nil, err
	}
	//解密pem格式的公钥
	pubBlock, _ := pem.Decode(pubKeyBytes)
	if pubBlock == nil {
		panic(errors.New("public key error"))
	}
	// 解析公钥
	pubInterface, err := x509.ParsePKIXPublicKey(pubBlock.Bytes)
	if err != nil {
		return nil, err
	}
	// 类型断言
	pub := pubInterface.(*rsa.PublicKey)
	// 解密
	partLen := pub.N.BitLen() / 8
	chunks := splitRSAChunks(ciphertext, partLen)
	buffer := bytes.NewBufferString("")
	for _, chunk := range chunks {
		decrypted, err := rsa.DecryptPKCS1v15(rand.Reader, priv, chunk)
		if err != nil {
			return nil, err
		}
		buffer.Write(decrypted)
	}

	return buffer.Bytes(), nil
}

func splitRSAChunks(buf []byte, lim int) [][]byte {
	var chunk []byte
	chunks := make([][]byte, 0, len(buf)/lim+1)
	for len(buf) >= lim {
		chunk, buf = buf[:lim], buf[lim:]
		chunks = append(chunks, chunk)
	}
	if len(buf) > 0 {
		chunks = append(chunks, buf[:])
	}
	return chunks
}
