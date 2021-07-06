package main

import "testing"

func TestVersionAllBlank(t *testing.T) {
	v := Version("", "", "", "")

	if v != "" {
		t.Error("Not blank", v)
	}
}

func TestVersionOnlyVersion(t *testing.T) {
	v := Version("12.19.3", "", "", "")

	expected := "12.19.3"
	if v != expected {
		t.Error("Not expected", expected, v)
	}
}

func TestVersionVersionBuilt(t *testing.T) {
	v := Version("12.19.3", "", "", "jim")

	expected := "12.19.3 [jim]"
	if v != expected {
		t.Error("Not expected", expected, v)
	}
}

func TestVersionAll(t *testing.T) {
	v := Version("12.19.3", "1234567890", "12/12/1900", "jim")

	expected := "12.19.3 1234567890 12/12/1900 [jim]"
	if v != expected {
		t.Error("Not expected", expected, v)
	}
}
