package util

import (
	"testing"
)

var testsGood = []struct {
	input string
	typ   string
	name  string
	num   int
}{
	{"eris_service_mint_1", "service", "mint", 1},
	{"eris_service_mint_10", "service", "mint", 10},
	{"eris_service_mint_love_10", "service", "mint_love", 10},
	{"eris_service_mint_loves_life_5", "service", "mint_loves_life", 5},
	{"eris_data_mint_1", "data", "mint", 1},
	{"eris_chain_mint_100", "chain", "mint", 100},
}

var testsBad = []string{
	"/noteris_service_mint_1",
	"/eris_service_mint_tnim_ecivres_sire",
	"noteris_service_mint_1",
	"eris_service_mint_tnim_ecivres_sire",
	"erisservicemint1234",
	"erisservice_mint_1234",
}

func TestContainerNameGood(t *testing.T) {
	for _, rt := range testsGood {
		c := ContainerDisassemble(rt.input)

		if c.ShortName != rt.name {
			t.Fatalf("Wrong shortname from %s. Got %s, expected %s", rt.input, c.ShortName, rt.name)
		}
		if c.Type != rt.typ {
			t.Fatalf("Wrong type from %s. Got %s, expected %s", rt.input, c.Type, rt.typ)
		}
		if c.Number != rt.num {
			t.Fatalf("Wrong number from %s. Got %s, expected %s", rt.input, c.Number, rt.num)
		}

		d := ContainerAssemble(rt.typ, rt.name, rt.num)

		if d.FullName != rt.input {
			t.Fatalf("Wrong full name from %s. Got %s, expected %s", rt.input, d.FullName, rt.input)
		}
	}
}

func TestContainerNameBad(t *testing.T) {
	for _, rt := range testsBad {

		d := ContainerDisassemble(rt)

		if d.FullName == rt {
			t.Fatalf("Unexpected return from %s. Got %s, expected nil", rt, d.FullName)
		}
	}
}
