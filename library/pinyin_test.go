package library

import (
	"log"
	"testing"
)

func TestGetPinyin(t *testing.T) {
	result := GetPinyin("Electronic Water Bath")

	log.Println(result)
}
