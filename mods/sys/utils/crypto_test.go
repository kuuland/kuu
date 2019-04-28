package utils

import (
	"fmt"
	"testing"
)

func TestMD5(t *testing.T) {
	fmt.Println(MD5("kuu"))
}

func TestGenerateFromPassword(t *testing.T) {
	fmt.Println(GenerateFromPassword("82d2934b974f475760420ddc79d47eb2"))
}

func TestCompareHashAndPassword(t *testing.T) {
	fmt.Println(CompareHashAndPassword("$2a$10$LjDBe919rpJ.WqIwT9rIGuGIPAfB8VuxcIJg6fgXFMHalf/zEt1f6", "82d2934b974f475760420ddc79d47eb2"))
}

func TestBase64Encode(t *testing.T) {
	fmt.Println(Base64Encode("hello kuu"))
}

func TestBase64Decode(t *testing.T) {
	fmt.Println(Base64Decode("aGVsbG8ga3V1"))
}
