package util

import (
	"reflect"
	"testing"
)

type S struct {
	String string
	Bool   bool
	Int    int
	Float  float64
	Map    map[string]string
	Slice  []string
}

var BasicTests = []struct {
	name string
	base *S
	over *S
	want *S
}{
	{"s1", &S{}, &S{}, &S{}},
	{"s2", &S{String: "a"}, &S{}, &S{String: "a"}},
	{"s3", &S{}, &S{String: "a"}, &S{String: "a"}},
	{"s4", &S{String: "a"}, &S{String: "b"}, &S{String: "b"}},

	{"b1", &S{}, &S{}, &S{}},
	{"b2", &S{Bool: true}, &S{}, &S{Bool: true}},
	{"b3", &S{}, &S{Bool: true}, &S{Bool: true}},
	{"b4", &S{Bool: true}, &S{Bool: false}, &S{Bool: true}},
	{"b5", &S{Bool: false}, &S{Bool: true}, &S{Bool: true}},

	{"i1", &S{}, &S{}, &S{}},
	{"i2", &S{Int: 10}, &S{}, &S{Int: 10}},
	{"i3", &S{}, &S{Int: 10}, &S{Int: 10}},
	{"i4", &S{Int: 10}, &S{Int: 12}, &S{Int: 12}},

	{"f1", &S{}, &S{}, &S{}},
	{"f2", &S{Float: 10.0}, &S{}, &S{Float: 10.0}},
	{"f3", &S{}, &S{Float: 10.0}, &S{Float: 10.0}},
	{"f4", &S{Float: 10.0}, &S{Float: 12.0}, &S{Float: 12.0}},

	{"m1",
		&S{Map: nil},
		&S{Map: nil},
		&S{Map: nil},
	},
	{"m2",
		&S{Map: nil},
		&S{Map: map[string]string{"a": "1"}},
		&S{Map: map[string]string{"a": "1"}},
	},
	{"m3",
		&S{Map: map[string]string{"a": "1"}},
		&S{Map: nil},
		&S{Map: map[string]string{"a": "1"}},
	},
	{"m4",
		&S{Map: map[string]string{}},
		&S{Map: map[string]string{}},
		&S{Map: map[string]string{}},
	},
	{"m5",
		&S{Map: map[string]string{}},
		&S{Map: map[string]string{"a": "1", "b": "2"}},
		&S{Map: map[string]string{"a": "1", "b": "2"}},
	},
	{"m6",
		&S{Map: map[string]string{"a": "1", "b": "2"}},
		&S{Map: map[string]string{}},
		&S{Map: map[string]string{"a": "1", "b": "2"}},
	},
	{"m7",
		&S{Map: map[string]string{"a": "1", "b": "2"}},
		&S{Map: map[string]string{"a": "1", "b": "2"}},
		&S{Map: map[string]string{"a": "1", "b": "2"}},
	},
	{"m8",
		&S{Map: map[string]string{"a": "1", "b": "2"}},
		&S{Map: map[string]string{"c": "3", "a": "4"}},
		&S{Map: map[string]string{"a": "4", "b": "2", "c": "3"}},
	},

	{"sl1",
		&S{Slice: []string{}},
		&S{Slice: []string{}},
		&S{Slice: []string{}},
	},
	{"sl2",
		&S{Slice: nil},
		&S{Slice: nil},
		&S{Slice: nil},
	},
	{"sl3",
		&S{Slice: nil},
		&S{Slice: []string{}},
		&S{Slice: []string{}},
	},
	{"sl4",
		&S{Slice: []string{}},
		&S{Slice: nil},
		&S{Slice: []string{}},
	},
	{"sl5",
		&S{Slice: nil},
		&S{Slice: []string{"a"}},
		&S{Slice: []string{"a"}},
	},
	{"sl6",
		&S{Slice: []string{"a"}},
		&S{Slice: nil},
		&S{Slice: []string{"a"}},
	},
	{"sl7",
		&S{Slice: []string{}},
		&S{Slice: []string{"1", "2"}},
		&S{Slice: []string{"1", "2"}},
	},
	{"sl8",
		&S{Slice: []string{"1", "2"}},
		&S{Slice: []string{}},
		&S{Slice: []string{"1", "2"}},
	},
	{"sl9",
		&S{Slice: []string{"1", "2"}},
		&S{Slice: []string{"1", "2"}},
		&S{Slice: []string{"1", "2", "1", "2"}},
	},
	{"sl10",
		&S{Slice: []string{"2", "1"}},
		&S{Slice: []string{"2", "1"}},
		&S{Slice: []string{"2", "1", "2", "1"}},
	},

	{"mix1",
		&S{Bool: true, Int: 10, Map: map[string]string{"a": "1"}},
		&S{Float: 12, Slice: []string{"2", "1"}},
		&S{Bool: true, Int: 10, Float: 12,
			Slice: []string{"2", "1"}, Map: map[string]string{"a": "1"}},
	},
	{"mix2",
		&S{Map: map[string]string{"a": "1"}, Slice: []string{"1"}},
		&S{Map: map[string]string{"a": "2"}, Slice: []string{"1"}, String: "a"},
		&S{Map: map[string]string{"a": "2"}, Slice: []string{"1", "1"}, String: "a"},
	},
}

func TestMergeBasic(t *testing.T) {
	for _, test := range BasicTests {
		if err := Merge(test.base, test.over); err != nil {
			t.Fatalf("%q: expected %v, got error %v", test.name, test.want, err)
		}
		if !reflect.DeepEqual(test.base, test.want) {
			t.Errorf("%q: expected %v, got %v", test.name, test.want, test.base)
		}
	}
}

func TestMergeError(t *testing.T) {
	if err := Merge(nil, nil); err != ErrMergeParameters {
		t.Fatalf("e1: expected error, got %v", err)
	}
	if err := Merge(S{}, nil); err != ErrMergeParameters {
		t.Fatalf("e2: expected error, got %v", err)
	}
	if err := Merge(nil, S{}); err != ErrMergeParameters {
		t.Fatalf("e3: expected error, got %v", err)
	}
	if err := Merge(S{}, S{}); err != ErrMergeParameters {
		t.Fatalf("e4: expected error, got %v", err)
	}
	if err := Merge(&S{}, "a"); err != ErrMergeParameters {
		t.Fatalf("e5: expected error, got %v", err)
	}
	if err := Merge("a", &S{}); err != ErrMergeParameters {
		t.Fatalf("e6: expected error, got %v", err)
	}
	if err := Merge(&struct{ A string }{A: "a"}, &S{}); err != ErrMergeParameters {
		t.Fatalf("e7: expected error, got %v", err)
	}
	if err := Merge(&S{}, &struct{ A string }{A: "a"}); err != ErrMergeParameters {
		t.Fatalf("e8: expected error, got %v", err)
	}
	if err := Merge(&[]int{}, &[]int{}); err != ErrMergeParameters {
		t.Fatalf("e9: expected error, got %v", err)
	}
	if err := Merge(&[]int{}, &S{}); err != ErrMergeParameters {
		t.Fatalf("e10: expected error, got %v", err)
	}
	if err := Merge(&S{}, &[]int{}); err != ErrMergeParameters {
		t.Fatalf("e11: expected error, got %v", err)
	}
}
