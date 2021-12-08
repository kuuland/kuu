package intl

import (
	"fmt"
	"testing"
	"time"
)

func TestFormatMessage(t *testing.T) {
	for i := 0; i < 100000; i++ {
		r := FormatMessage(
			fmt.Sprintf("你好 {{ name%d }}, 欢迎{{    name%d   }}，现在时间{{now}}", i, i),
			map[string]interface{}{
				fmt.Sprintf("name%d", i): "Daniel",
				"now":                    time.Now().Format(time.RFC3339),
			},
		)
		t.Log(r)
	}
}
