package maker

import (
	"testing"
)

func TestConvertExportedSliceToString(t *testing.T) {
	if res := convertExportPortsSliceToString([]string{}); res != "" {
		t.Errorf("Failed to convert empty slice: %s", res)
	}
	if res := convertExportPortsSliceToString([]string{"1"}); res != `[ "1" ]` {
		t.Errorf("Failed to convert singleton slice: %s", res)
	}
	if res := convertExportPortsSliceToString([]string{"1", "3"}); res != `[ "1", "3" ]` {
		t.Errorf("Failed to convert slice: %s", res)
	}
}
