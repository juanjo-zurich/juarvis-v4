package output

import (
	"testing"
)

func TestSetJSONMode(t *testing.T) {
	SetJSONMode(true)
	if !IsJSONMode() {
		t.Error("expected JSON mode to be true")
	}
	SetJSONMode(false)
	if IsJSONMode() {
		t.Error("expected JSON mode to be false")
	}
}

func TestSuccess_NoPanic(t *testing.T) {
	SetJSONMode(false)
	Success("test %s", "message")
}

func TestError_NoPanic(t *testing.T) {
	SetJSONMode(false)
	Error("test %s", "error")
}

func TestWarning_NoPanic(t *testing.T) {
	SetJSONMode(false)
	Warning("test %s", "warning")
}

func TestInfo_NoPanic(t *testing.T) {
	SetJSONMode(false)
	Info("test %s", "info")
}

func TestPrintTable_NoPanic(t *testing.T) {
	SetJSONMode(false)
	PrintTable([]string{"A", "B"}, [][]string{{"1", "2"}, {"3", "4"}})
}

func TestPrintJSON_NoPanic(t *testing.T) {
	SetJSONMode(true)
	PrintJSON(map[string]string{"key": "value"})
	SetJSONMode(false)
}

func TestSuccess_JSONMode(t *testing.T) {
	SetJSONMode(true)
	Success("test")
	SetJSONMode(false)
}

func TestError_JSONMode(t *testing.T) {
	SetJSONMode(true)
	Error("test error")
	SetJSONMode(false)
}

func TestPrintTable_JSONMode(t *testing.T) {
	SetJSONMode(true)
	PrintTable([]string{"A", "B"}, [][]string{{"1", "2"}})
	SetJSONMode(false)
}
