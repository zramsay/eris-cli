package util

import "testing"

var Tests = []struct {
	v1   string
	v2   string
	want bool
}{
	{"1.0", "1.0", true},
	{"1.1", "1.0", true},
	{"1.0", "1.1", false},
	{"1.6", "1.7", false},
	{"10.7", "10.8", false},
	{"1.8", "1.8", true},
	{"10.9", "10.8", true},
	{"1.10", "1.9", true},
	{"2.0", "1.8", true},
	{"1.8", "2.0", false},
	{"10.8", "20.0", false},

	{"1.0.0", "1.0", true},
	{"1.0", "1.0.0", true},
	{"1.0.1", "1.0", true},
	{"1.0", "1.0.1", true},
	{"0.4.1", "0.3.9", true},
	{"0.3.9", "0.4.1", false},

	// Unacceptable values.
	{"0", "0", false},
	{"0", "", false},
	{"", "0", false},
	{"", "", false},
	{"9", "0", false},
	{"0", "9", false},
	{"1", "2", false},
	{"3", "4", false},
	{"b.1", "d.0", false},
	{"ge.ge.", ".ge.ge", false},
	{"1.0", "b.0", false},
	{"9.0", "+.-", false},
	{"3.4", "3.*", false},
	{"4.##", "1.1", false},
}

func TestCompareVersions(t *testing.T) {
	for _, test := range Tests {
		if actual := CompareVersions(test.v1, test.v2); actual != test.want {
			t.Errorf("expected %v comparing %v with %v", test.want, test.v1, test.v2)
		}
	}
}
