package intl

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLanguages(t *testing.T) {
	assert.Equal(t, "简体中文", LanguageMap()["zh-Hans"])
}

func TestLanguageList(t *testing.T) {
	list := LanguageList()
	assert.Equal(t, "English", list[2].Name)
	assert.Equal(t, "ja", list[7].Code)
}
