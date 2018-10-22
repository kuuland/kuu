package mongo

import (
	"fmt"
	"github.com/kuuland/kuu/plugins/mongo/db"
	"testing"
)

var uri = "mongodb://root:kuuland@127.0.0.1:27017/kuu?authSource=admin&maxPoolSize=50"

func TestConnect(t *testing.T) {
	db.Connect(uri)
	if n, err := C("user").Count(); err == nil {
		fmt.Println(n)
	} else {
		fmt.Println(err)
	}
}
