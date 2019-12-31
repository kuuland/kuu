package kuu

import (
	"fmt"
	"testing"
)

func TestGenRSAKey(t *testing.T) {
	prvKey, pubKey := GenRSAKey()
	t.Log(string(prvKey), string(pubKey))
}

func TestRSA(t *testing.T) {
	// 加密
	privateKey := []byte(`-----BEGIN RSA PRIVATE KEY-----
MIICXAIBAAKBgQCs7c6USBv10SNpzydjeREzmp7iepJQDiEcRiSCPstfDwvJ6pVO
X4tXrNcJwnlfEuU7kgKVUmWQXbtBl5cIzAbaZgG5fOqou8ZxD39/JYYb2P3XtDTf
KnXXaqB+VVdGm6MDdNDAuCDAPboj3QSDZ6OEBLaVwQc6N7PZgyCtIBcPoQIDAQAB
AoGAT8xFAYPs8xgZAWCISoy5dViqbNQm5C5A9S0g98FGU4074WcQkuPgBwtJB8Xo
AAlWIpEUBBfLqjy2hmQPXA3aMvbhs2yHksIUz0g1f1oEjJa1TB5qImYAIr7TqXVm
Gc+ZrXsHUOhFFzPJ03Bzn5azlkEnA3KPh3CJJt71cns8IYECQQDl2xr7iWadF1+I
/mOQ8NU/D0+QAbg3tzzL0EmsyEz93pn+2g8GAEtW01D8f7BvrEwGlgkVh7R+e3AI
utw5cL05AkEAwJkdZP8jMQXXMU4z4RL6BAzyvFk0QQ8EAIfGvUA0vRvyxluhxKGj
RBwtULSiZLb1/i8jXknjR1fw3f6qacBNqQJAe8WCQBR61vhxDzm8r52flrdN5oOm
iQn4iN997LZnDwVA80TEdjzOVNCxeWXgwiGLRrif56INhVY+u9SzJZMZsQJBAK8c
B4/GMXbm+orHsX+YQ1zfcOsyp8HnJxpcWKPE9q5h9M/IjEI9PDY28DSKp4OuneYn
cZ7OyygYmtUcMFDKGVECQCflOz8XfZcaA5VU+OS7Seu6QprTNjR4rHCKLsVhWvjM
vxcYBAyIbAoilS7F85k+k9aUgk/9Ng8QeFQN5eC7Pnc=
-----END RSA PRIVATE KEY-----
`)
	publicKey := []byte(`-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCs7c6USBv10SNpzydjeREzmp7i
epJQDiEcRiSCPstfDwvJ6pVOX4tXrNcJwnlfEuU7kgKVUmWQXbtBl5cIzAbaZgG5
fOqou8ZxD39/JYYb2P3XtDTfKnXXaqB+VVdGm6MDdNDAuCDAPboj3QSDZ6OEBLaV
wQc6N7PZgyCtIBcPoQIDAQAB
-----END PUBLIC KEY-----
`)
	originalData := []byte(`Kubernetes 是用于自动部署，扩展和管理容器化应用程序的开源系统。
它将组成应用程序的容器组合成逻辑单元，以便于管理和服务发现。Kubernetes 源自Google 15 年生产环境的运维经验，同时凝聚了社区的最佳创意和实践。
Google 每周运行数十亿个容器，Kubernetes 基于与之相同的原则来设计，能够在不扩张运维团队的情况下进行规模扩展。
无论是本地测试，还是跨国公司，Kubernetes 的灵活性都能让你在应对复杂系统时得心应手。
Kubernetes 是开源系统，可以自由地部署在企业内部，私有云、混合云或公有云，让您轻松地做出合适的选择。`)
	encryptedData, err := RSAEncrypt(originalData, publicKey)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(fmt.Sprintf("Encrypted Data: %s", string(encryptedData)))
	// 签名
	sign, err := RSASignWithSha256(originalData, privateKey)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(fmt.Sprintf("Signature: %s", sign))
	// 解密
	decryptedData, err := RSADecrypt(encryptedData, publicKey, privateKey)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(fmt.Sprintf("Decrypted Data: %s", string(decryptedData)))
	// 验签
	err = RSAVerySignWithSha256(decryptedData, sign, publicKey)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log("Signature verification passed")
}
