package kuu

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"io"
	"io/ioutil"

	"github.com/gin-gonic/gin"
)

// Join 基于字节实现的字符串拼接
func Join(args ...string) string {
	b := bytes.Buffer{}
	for _, item := range args {
		b.WriteString(item)
	}
	return b.String()
}

// CloneDeep 深度拷贝
func CloneDeep(src, dst interface{}) error {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(src); err != nil {
		return err
	}
	return gob.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(dst)
}

// CopyBody 复制请求体
func CopyBody(c *gin.Context, docs interface{}) (err error) {
	var dst io.Reader
	if err = CloneDeep(c.Request.Body, dst); err == nil {
		var data []byte
		if data, err = ioutil.ReadAll(dst); err == nil {
			json.Unmarshal(data, docs)
		}
	}
	return err
}
