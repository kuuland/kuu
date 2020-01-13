package kuu

import (
	"reflect"
	"testing"
)

// TestRandCode
func TestRandCode(t *testing.T) {
	t.Log(RandCode(4))
	t.Log(RandCode(6))
	t.Log(RandCode())
	t.Log(RandCode(10))
}

func TestSetUrlQuery(t *testing.T) {
	var u string

	u = "/mp"
	t.Log(SetUrlQuery(u, map[string]interface{}{"t": 1, "s": "hello"}))

	u = "/mp?a=1&b=2"
	t.Log(SetUrlQuery(u, map[string]interface{}{"t": 1, "s": "hello"}))
	t.Log(SetUrlQuery(u, map[string]interface{}{"t": 1, "s": "hello"}, true))

	u = "https://www.example.com/mp"
	t.Log(SetUrlQuery(u, map[string]interface{}{"t": 1, "s": "hello"}))

	u = "https://www.example.com/mp?a=1&b=2"
	t.Log(SetUrlQuery(u, map[string]interface{}{"t": 1, "s": "hello"}))
	t.Log(SetUrlQuery(u, map[string]interface{}{"t": 1, "s": "hello"}, true))

	u = "https://www.example.com/admin/?a=1&b=2#/test?h=1&j=ok"
	t.Log(SetUrlQuery(u, map[string]interface{}{"t": 1, "s": "hello"}))
	t.Log(SetUrlQuery(u, map[string]interface{}{"t": 1, "s": "hello"}, true))
}

func TestParseJSONPath(t *testing.T) {
	tests := []struct {
		name string
		path string
		want []string
	}{
		{
			name: "name[0]",
			path: "name[0]",
			want: []string{"name", "[0]"},
		},
		{
			name: "name[192]",
			path: "name[192]",
			want: []string{"name", "[192]"},
		},
		{
			name: "name[1][0].[2]",
			path: "name[1][0].[2]",
			want: []string{"name", "[1]", "[0]", "[2]"},
		},
		{
			name: "a.b.c.name[1].e.f[2].t[3]",
			path: "a.b.c.name[1].e.f[2].t[3]",
			want: []string{"a", "b", "c", "name", "[1]", "e", "f", "[2]", "t", "[3]"},
		},
		{
			name: "[0][100][398][89]name[1][2][3]",
			path: "[0][100][398][89]name[1][2][3]",
			want: []string{"[0]", "[100]", "[398]", "[89]", "name", "[1]", "[2]", "[3]"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseJSONPath(tt.path); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseKeys() = %v, want %v", got, tt.want)
			}
		})
	}
}
