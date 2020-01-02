package kuu

import (
	"fmt"
	"testing"
)

func TestAESCBC(t *testing.T) {
	key := []byte("hgfedcba87654321")
	raw := []byte(`Kubernetes 是用于自动部署，扩展和管理容器化应用程序的开源系统。它将组成应用程序的容器组合成逻辑单元，以便于管理和服务发现。Kubernetes 源自Google 15 年生产环境的运维经验，同时凝聚了社区的最佳创意和实践。`)

	// Encrypt
	encryptedData, err := AESCBCEncrypt(raw, key)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(fmt.Sprintf("Encrypted Data: %s", encryptedData))

	// Decrypt
	decryptedData, err := AESCBCDecrypt(encryptedData, key)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(fmt.Sprintf("Decrypted Data: %s", decryptedData))
}
