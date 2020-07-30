package main

import "testing"

func TestSliceRemoveString(t *testing.T) {
	s := []string{"a", "b", "c", "d"}
	x := "c"

	o, err := SliceRemoveString(s, x)
	if err != nil {
		t.Error(err)
	}

	e := []string{"a", "b", "d"}
	if len(o) != len(e) {
		t.Error("ouptut array of incorrect length")
	}
	for i := range e {
		if o[i] != e[i] {
			t.Error("incorrect result")
		}
	}

}
