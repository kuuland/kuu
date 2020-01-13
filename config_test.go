package kuu

import (
	"reflect"
	"testing"
)

func TestConfig_ParseKeys(t *testing.T) {
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
			c := &Config{}
			if got := c.ParseKeys(tt.path); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseKeys() = %v, want %v", got, tt.want)
			}
		})
	}
}
