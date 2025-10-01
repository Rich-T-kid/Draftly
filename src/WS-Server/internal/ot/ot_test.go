package internal

import "testing"

func TestWork(t *testing.T) {
	var a = 2
	if a != 2 {
		t.Errorf("Expected 2, got %d", a)
	}
}
