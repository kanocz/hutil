package hutil

import (
	"fmt"
	"testing"
)

func Test_IsLangID(t *testing.T) {

	data := []struct {
		lang   string
		result bool
	}{
		{"ru", true},
		{"en", true},
		{"cz", true},
		{"c1", false},
		{"ру", false},
		{",c", false},
		{"", false},
		{"c", false},
		{"czc", false},
	}

	for _, d := range data {
		if IsLangID(d.lang) != d.result {
			t.Error(fmt.Sprintf("IsLangID did not work as expected. IsLangID(%s) is %v but has to be %v", d.lang, IsLangID(d.lang), d.result))
		}
	}
}

func Test_IsHexString(t *testing.T) {
	data := []struct {
		lang   string
		result bool
	}{
		{"131", true},
		{"f2e4", true},
		{"f2-e4", false},
		{"c", true},
		{"131.", false},
		{"рус1", false},
		{"", false},
		{"32_", false},
		{" ", false},
		{" 3e4j", false},
	}

	for _, d := range data {
		if IsHexString(d.lang) != d.result {
			t.Error(fmt.Sprintf("IsHexString did not work as expected. IsHexString(%s) is %v but has to be %v", d.lang, IsHexString(d.lang), d.result))
		}
	}
}

func Test_IsUUID(t *testing.T) {
	data := []struct {
		lang   string
		result bool
	}{
		{"131", true},
		{"f2e4", true},
		{"f2-e4", true},
		{"f2-e4-34f", true},
		{"c", true},
		{"131.", false},
		{"рус1", false},
		{"", false},
		{"32_", false},
		{" ", false},
		{" 3e4j", false},
	}

	for _, d := range data {
		if IsUUID(d.lang) != d.result {
			t.Error(fmt.Sprintf("IsUUID did not work as expected. IsUUID(%s) is %v but has to be %v", d.lang, IsUUID(d.lang), d.result))
		}
	}

}
