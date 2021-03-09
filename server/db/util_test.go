package db

import (
	"testing"
)

func TestSetPinyin(t *testing.T) {
	if p := SetPinyin("ni3 hao3"); p != "ni[3] hao[3]" {
		t.Error(p + " outputted")
	}

	if p := SetPinyin("nǐ hǎo"); p != "ni[3] hao[3]" {
		t.Error(p + " outputted")
	}
}

func TestMakePinyin(t *testing.T) {
	if p := MakePinyin("ni[3] hao[3]"); p != "ni3 hao3" {
		t.Error(p + " outputted")
	}
}
