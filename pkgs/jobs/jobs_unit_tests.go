package jobs

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/monax/cli/log"
	"github.com/monax/cli/testutil"
)

func TestMain(m *testing.M) {
	log.SetLevel(log.ErrorLevel)
	testutil.IfExit(testutil.Init(testutil.Pull{
		Images: []string{"data", "db", "keys", "ipfs"},
	}))
	exitCode := m.Run()
	testutil.IfExit(testutil.TearDown())
	os.Exit(exitCode)
}

func TestEPMTypeAssertions(t *testing.T) {
	jobs := &Jobs{
		Account: "",
		Jobs:    []*Job{},
		JobMap:  nil,
	}
	for i, test := range []struct {
		key      interface{} //jobs.Type for type Type
		val      interface{}
		relation string
		procErr  string
		execErr  string
	}{
		// Strings Valid
		{
			Type{
				StringResult: "hello",
				ActualResult: "hello",
			},
			"hello",
			"eq",
			"",
			"",
		},
		{
			"hello",
			"hello",
			"eq",
			"",
			"",
		},
		{
			Type{
				StringResult: "hello",
				ActualResult: "hello",
			},
			Type{
				StringResult: "hello",
				ActualResult: "hello",
			},
			"eq",
			"",
			"",
		},
		{
			Type{
				StringResult: "hello",
				ActualResult: "hello",
			},
			Type{
				StringResult: "hello\n",
				ActualResult: "hello\n",
			},
			"eq",
			"",
			"",
		},
		{
			Type{
				StringResult: "01234567890123456789",
				ActualResult: "01234567890123456789",
			},
			Type{
				StringResult: "1234567890123456789",
				ActualResult: 1234567890123456789,
			},
			"eq",
			"",
			"",
		},
		//Strings invalid
		{
			Type{
				StringResult: "01234567890123456789",
				ActualResult: "01234567890123456789",
			},
			Type{
				StringResult: "0x01234567890123456789",
				ActualResult: "0x01234567890123456789",
			},
			"eq",
			"",
			"Assertion Failed!",
		},
		{
			Type{
				StringResult: "01234567890123456789",
				ActualResult: "01234567890123456789",
			},
			false,
			"!=",
			"Assertion Failed!: Cannot convert key type string to val type bool",
			"",
		},
		{
			Type{
				StringResult: "01234567890123456789",
				ActualResult: "01234567890123456789",
			},
			[]interface{}{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
			"!=",
			"Assertion Failed!: Cannot convert key type string to val type []interface {}",
			"",
		},
		{
			Type{
				StringResult: "01234567890123456789",
				ActualResult: "01234567890123456789",
			},
			Type{
				StringResult: "1234567890123456789",
				ActualResult: 3234567890123456789,
			},
			">=",
			"",
			"Assertion Failed!",
		},
		{
			Type{
				StringResult: "01234567890123456789",
				ActualResult: "01234567890123456789",
			},
			Type{
				StringResult: "false",
				ActualResult: false,
			},
			"!=",
			"Assertion Failed!: Cannot convert key type string to val type bool",
			"",
		},
		// Bools
		{
			true,
			true,
			"eq",
			"",
			"",
		},
		{
			Type{
				StringResult: "false",
				ActualResult: false,
			},
			Type{
				StringResult: "false",
				ActualResult: "false",
			},
			"eq",
			"",
			"",
		},
		{
			Type{
				StringResult: "false",
				ActualResult: false,
			},
			Type{
				StringResult: "0",
				ActualResult: false,
			},
			"eq",
			"",
			"",
		},
		{
			Type{
				StringResult: "false",
				ActualResult: false,
			},
			Type{
				StringResult: "false",
				ActualResult: 0,
			},
			"eq",
			"",
			"",
		},
		{
			Type{
				StringResult: "false",
				ActualResult: false,
			},
			Type{
				StringResult: "0",
				ActualResult: []byte{0}, //same as uint8
			},
			"eq",
			"",
			"",
		},
		{
			Type{
				StringResult: "false",
				ActualResult: false,
			},
			Type{
				StringResult: "0",
				ActualResult: int(0), //same as false
			},
			"eq",
			"",
			"",
		},
		{
			Type{
				StringResult: "true",
				ActualResult: true,
			},
			Type{
				StringResult: "1",
				ActualResult: int(1), //same as false
			},
			"eq",
			"",
			"",
		},
		{
			Type{
				StringResult: "true",
				ActualResult: true,
			},
			Type{
				StringResult: "1",
				ActualResult: []byte{1}, //same as false
			},
			"eq",
			"",
			"",
		},
		{
			Type{
				StringResult: "true",
				ActualResult: true,
			},
			Type{
				StringResult: "true",
				ActualResult: true, //same as false
			},
			"eq",
			"",
			"",
		},
		// Bools Invalid
		{
			true,
			false,
			"==",
			"",
			"Assertion Failed!",
		},
		{
			Type{
				StringResult: "true",
				ActualResult: true,
			},
			Type{
				StringResult: "false",
				ActualResult: "false",
			},
			"==",
			"",
			"Assertion Failed!",
		},
		{
			Type{
				StringResult: "true",
				ActualResult: true,
			},
			Type{
				StringResult: "[true]",
				ActualResult: []interface{}{true},
			},
			"!=",
			"Assertion Failed!: Cannot convert key type bool to val type []interface {}",
			"",
		},
		{
			Type{
				StringResult: "true",
				ActualResult: true,
			},
			Type{
				StringResult: "0",
				ActualResult: false,
			},
			"!=",
			"",
			"",
		},
		{
			Type{
				StringResult: "true",
				ActualResult: true,
			},
			Type{
				StringResult: "false",
				ActualResult: 0,
			},
			"!=",
			"",
			"",
		},
		{
			Type{
				StringResult: "true",
				ActualResult: true,
			},
			Type{
				StringResult: "0",
				ActualResult: []byte{0}, //same as uint8
			},
			"!=",
			"",
			"",
		},
		{
			Type{
				StringResult: "true",
				ActualResult: true,
			},
			Type{
				StringResult: "0",
				ActualResult: int(0), //same as false
			},
			"!=",
			"",
			"",
		},
		// integers valid
		{
			1,
			1,
			"eq",
			"",
			"",
		},
		{
			1,
			true,
			"eq",
			"",
			"",
		},
		{
			0,
			false,
			"eq",
			"",
			"",
		},
		// integers invalid
		{
			Type{
				StringResult: "0x1020509",
				ActualResult: 0x1020509,
			},
			Type{
				StringResult: "0x1020509",
				ActualResult: "0x1020509",
			},
			"!=",
			"",
			"",
		},
		{
			Type{
				StringResult: "2",
				ActualResult: 2,
			},
			false,
			"!=",
			"Assertion Failed!: Cannot convert key type int to val type bool",
			"",
		},
		// string slices valid
		{
			Type{
				StringResult: `["hello", "world"]`,
				ActualResult: []interface{}{"hello", "world"},
			},
			Type{
				StringResult: `["hello", "world"]`,
				ActualResult: []interface{}{"hello", "world"},
			},
			"eq",
			"",
			"",
		},
		{
			Type{
				StringResult: `["hello", "world"]`,
				ActualResult: []interface{}{"hello", "world"},
			},
			Type{
				StringResult: `["hello", "world"]`,
				ActualResult: `["hello", "world"]`,
			},
			"eq",
			"",
			"",
		},
		{
			Type{
				StringResult: `["hello", "world"]`,
				ActualResult: []interface{}{"hello", "world"},
			},
			[]interface{}{"hello", "world"},
			"eq",
			"",
			"",
		},
		// string slices invalid
		{
			Type{
				StringResult: `["hello", "world"]`,
				ActualResult: []interface{}{"hello", "world"},
			},
			false,
			"!=",
			"Assertion Failed!: Cannot convert key type []interface {} to val type bool",
			"",
		},
		{
			Type{
				StringResult: `["hello", "world"]`,
				ActualResult: []interface{}{"hello", "world"},
			},
			12,
			"!=",
			"Assertion Failed!: Cannot convert key type []interface {} to val type int",
			"",
		},
		{
			Type{
				StringResult: `["hello", "world"]`,
				ActualResult: []interface{}{"hello", "world"},
			},
			"hello",
			"eq",
			"Assertion Failed!: Cannot convert key type []interface {} to val type string",
			"",
		},
		// bool slices valid
		{
			Type{
				StringResult: "[true, false]",
				ActualResult: []bool{true, false},
			},
			Type{
				StringResult: "[true, false]",
				ActualResult: []bool{true, false},
			},
			"eq",
			"",
			"",
		},
		{
			Type{
				StringResult: "[true, false]",
				ActualResult: []interface{}{true, false},
			},
			Type{
				StringResult: `[true, false]`,
				ActualResult: `[true, false]`,
			},
			"eq",
			"",
			"",
		},
		{
			Type{
				StringResult: "[true, false]",
				ActualResult: []interface{}{true, false},
			},
			[]interface{}{true, false},
			"eq",
			"",
			"",
		},
		// bool slices invalid
		{
			Type{
				StringResult: "[true, false]",
				ActualResult: []interface{}{true, false},
			},
			false,
			"!=",
			"Assertion Failed!: Cannot convert key type []interface {} to val type bool",
			"",
		},
		{
			Type{
				StringResult: "[true, false]",
				ActualResult: []interface{}{true, false},
			},
			12,
			"!=",
			"Assertion Failed!: Cannot convert key type []interface {} to val type int",
			"",
		},
		{
			Type{
				StringResult: "[true, false]",
				ActualResult: []interface{}{true, false},
			},
			"hello",
			"!=",
			"Assertion Failed!: Cannot convert key type []interface {} to val type string",
			"",
		},
		// integer slices valid
		{
			Type{
				StringResult: "[1, 2]",
				ActualResult: []interface{}{1, 2},
			},
			Type{
				StringResult: "[1, 2]",
				ActualResult: []interface{}{1, 2},
			},
			"eq",
			"",
			"",
		},
		{
			Type{
				StringResult: "[1, 2]",
				ActualResult: []interface{}{1, 2},
			},
			Type{
				StringResult: `[1, 2]`,
				ActualResult: `[1, 2]`,
			},
			"eq",
			"",
			"",
		},
		{
			Type{
				StringResult: "[1, 2]",
				ActualResult: []interface{}{1, 2},
			},
			[]interface{}{true, false},
			"eq",
			"",
			"Assertion Failed!",
		},
		// integer slices invalid
		{
			Type{
				StringResult: "[1, 2]",
				ActualResult: []interface{}{1, 2},
			},
			false,
			"!=",
			"Assertion Failed!: Cannot convert key type []interface {} to val type bool",
			"",
		},
		{
			Type{
				StringResult: "[1, 2]",
				ActualResult: []interface{}{1, 2},
			},
			12,
			"!=",
			"Assertion Failed!: Cannot convert key type []interface {} to val type int",
			"",
		},
		{
			Type{
				StringResult: "[1, 2]",
				ActualResult: []interface{}{1, 2},
			},
			"hello",
			"!=",
			"Assertion Failed!: Cannot convert key type []interface {} to val type string",
			"",
		},
	} {
		assert := &Assert{
			Key:      test.key,
			Relation: test.relation,
			Value:    test.val,
		}

		log.WithFields(log.Fields{
			"key=>":      test.key,
			"relation=>": test.relation,
			"value=>":    test.val,
		}).Error("Testing Assertion Job ", i, "\n")

		err := assert.PreProcess(jobs)
		if fullErr, ok := errorsProper(err, test.procErr); !ok {
			t.Error("Test ", i, ": ", fullErr)
			continue
		}
		if test.procErr == "" {
			_, err = assert.Execute(jobs)
			if fullErr, ok := errorsProper(err, test.execErr); !ok {
				t.Error("Test ", i, ": ", fullErr)
			}
		}
	}
}

func TestPreProcessingThroughSet(t *testing.T) {
	jobs := &Jobs{
		Account: "",
		Jobs:    []*Job{},
		JobMap: map[string]*JobResults{
			"String": {
				FullResult:   Type{"hello", "hello"},
				NamedResults: nil,
			},
			"Bool": {
				FullResult:   Type{"false", false},
				NamedResults: nil,
			},
			"Int": {
				FullResult:   Type{"1", 1},
				NamedResults: nil,
			},
			"StringSlice": {
				FullResult:   Type{"[hello, world]", []interface{}{"hello", "world"}},
				NamedResults: nil,
			},
			"BoolSlice": {
				FullResult:   Type{"[false, true]", []interface{}{false, true}},
				NamedResults: nil,
			},
			"IntSlice": {
				FullResult:   Type{"[1, 2]", []interface{}{1, 2}},
				NamedResults: nil,
			},
			"Mapping": {
				FullResult: Type{},
				NamedResults: map[string]Type{
					"String":      Type{"hello", "hello"},
					"Bool":        Type{"false", false},
					"Int":         Type{"1", 1},
					"StringSlice": Type{"[hello, world]", []interface{}{"hello", "world"}},
					"BoolSlice":   Type{"[false, true]", []interface{}{false, true}},
					"IntSlice":    Type{"[1, 2]", []interface{}{1, 2}},
				},
			},
		},
	}

	for i, test := range []struct {
		input          interface{} //jobs.Type for type Type
		expectedOutput Type
	}{
		//initial testing of set val preprocessing
		{
			"hello",
			Type{
				StringResult: "hello",
				ActualResult: "hello",
			},
		},
		{
			1,
			Type{"1", 1},
		},
		{
			false,
			Type{"false", false},
		},
		{
			[]interface{}{"hello", "world"},
			Type{`["hello","world"]`, []interface{}{"hello", "world"}},
		},
		{
			[]interface{}{1, 2},
			Type{"[1,2]", []interface{}{1, 2}},
		},
		{
			[]interface{}{false, true},
			Type{"[false,true]", []interface{}{false, true}},
		},
		// test regular preprocessing
		{
			"$String",
			Type{"hello", "hello"},
		},
		{
			"$Int",
			Type{"1", 1},
		},
		{
			"$Bool",
			Type{"false", false},
		},
		{
			"$StringSlice",
			Type{"[hello, world]", []interface{}{"hello", "world"}},
		},
		{
			"$IntSlice",
			Type{"[1, 2]", []interface{}{1, 2}},
		},
		{
			"$BoolSlice",
			Type{"[false, true]", []interface{}{false, true}},
		},
		// Test nesting values to be preprocessed
		{
			[]interface{}{"$Bool", "$String", "$Int", "$BoolSlice", "$IntSlice", "$StringSlice"},
			Type{
				`[false,"hello",1,[false,true],[1,2],["hello","world"]]`,
				[]interface{}{false, "hello", 1, []interface{}{false, true}, []interface{}{1, 2}, []interface{}{"hello", "world"}},
			},
		},
		// Test named result mappings
		{
			"$Mapping.String",
			Type{"hello", "hello"},
		},
		{
			"$Mapping.Int",
			Type{"1", 1},
		},
		{
			"$Mapping.Bool",
			Type{"false", false},
		},
		{
			"$Mapping.StringSlice",
			Type{"[hello, world]", []interface{}{"hello", "world"}},
		},
		{
			"$Mapping.IntSlice",
			Type{"[1, 2]", []interface{}{1, 2}},
		},
		{
			"$Mapping.BoolSlice",
			Type{"[false, true]", []interface{}{false, true}},
		},

		/* [rj] todo when I figure out docker linking to chains and when bonding jobs come back
		Question: For a set job should this resolve at time set job is referenced or when it is fired off?
		{
			"$block"
		}
		{
			"$block+1",
		},
		{
			"$block + 1",
		},
		{
			"$block-1",
		},
		{
			"$block - 1",
		},
		*/
	} {
		set := &Set{
			Value: test.input,
		}

		log.WithFields(log.Fields{
			"inputting =>":        test.input,
			"expecting output =>": test.expectedOutput,
		}).Error("Testing Set Job ", i)
		err := set.PreProcess(jobs)
		if err != nil {
			t.Fatalf("Unexpected Error: %v", err)
		}
		result, err := set.Execute(jobs)
		if err != nil {
			t.Fatalf("Unexpected Error: %v", err)
		}
		fullResult := result.FullResult
		if !reflect.DeepEqual(fullResult, test.expectedOutput) {
			t.Fatalf("Set/Preprocessing failed: Expected %v got %v", test.expectedOutput, fullResult)
		}
	}
}

func TestAccountJobs(t *testing.T) {
	jobs := &Jobs{
		Account: "",
		Jobs:    []*Job{},
		JobMap:  nil,
	}
	for i, test := range []struct {
		input          string
		expectedOutput string
		expectedErr    string
	}{
		{
			"73DF585F1F16912D0A4138BE7020F789C94B1F75",
			"73DF585F1F16912D0A4138BE7020F789C94B1F75",
			"",
		},
		{
			"73Df585f1f16912d0a4138BE7020f789C94b1F75",
			"73DF585F1F16912D0A4138BE7020F789C94B1F75",
			"",
		},
		{
			"0xC2c2c26961e5560081003Bb157549916B21744Db",
			"C2C2C26961E5560081003BB157549916B21744DB",
			"",
		},
	} {
		account := &Account{
			Address: test.input,
		}
		err := account.PreProcess(jobs)
		if fullErr, ok := errorsProper(err, test.expectedErr); !ok {
			t.Error("Test ", i, ": ", fullErr)
			continue
		}
		result, err := account.Execute(jobs)
		if !reflect.DeepEqual(result.FullResult.StringResult, test.expectedOutput) {
			t.Errorf("Expected %v, got %v", test.expectedOutput, result.FullResult.StringResult)
		}
	}
}

func TestQueryVals(t *testing.T) {
	jobs := &Jobs{
		Account: "",
		Jobs:    []*Job{},
		JobMap:  nil,
	}
	for i, test := range []struct {
		input       string
		expectedErr string
	}{
		{"bonded_validators", ""},
		{"unbonded_validators", ""},
		{"somethingElse", "Invalid value passed in, expected bonded_validators or unbonded_validators, got somethingElse"},
	} {
		qVals := &QueryVals{
			Field: test.input,
		}
		err := qVals.PreProcess(jobs)
		if fullErr, ok := errorsProper(err, test.expectedErr); !ok {
			t.Error("Test ", i, ": ", fullErr)
			continue
		}
	}
}

func errorsProper(err error, test string) (string, bool) {
	switch {
	case err != nil && len(test) == 0:
		return fmt.Sprintf("Expected no err but got: %v", err), false
	case err == nil && len(test) != 0:
		return fmt.Sprintf("Expected err: %v but got none", test), false
	case err != nil && len(test) != 0 && err.Error() != test:
		return fmt.Sprintf("Expected err: '%v' got err: '%v'", test, err), false
	default:
		return "", true
	}
}
