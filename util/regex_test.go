package util

import (
	"testing"
)

var regTestsGood = []struct {
	input string
	name  string
	num   string
}{
	{"/eris_service_mint_1", "mint", "1"},
	{"/eris_service_mint_10", "mint", "10"},
	{"/eris_service_mint_love_10", "mint_love", "10"},
	{"/eris_service_mint_loves_life_5", "mint_loves_life", "5"},
}

var regTestsBad = []string{
	"/noteris_service_mint_1",
	"/eris_service_mint_tnim_ecivres_sire",
}

func TestRegexGood(t *testing.T) {
	r := erisRegExp("service")

	for _, rt := range regTestsGood {
		match := r.FindAllStringSubmatch(rt.input, 1)
		if len(match) == 0 {
			t.Fatalf("Found no match for %s", rt.input)
		}
		m := match[0]
		if m[1] != rt.name {
			t.Fatalf("Wrong name from %s. Got %s, expected %s", rt.input, m[1], rt.name)
		}

		if m[2] != rt.num {
			t.Fatalf("Wrong number from %s. Got %s, expected %s", rt.input, m[2], rt.num)
		}
	}
}

func TestRegexBad(t *testing.T) {
	r := erisRegExp("service")

	for _, rt := range regTestsBad {
		match := r.FindAllStringSubmatch(rt, 1)
		if len(match) != 0 {
			t.Fatalf("Found match for %s when we should not have: %v", rt, match)
		}
	}
}
