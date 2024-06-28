package common

import "testing"

func TestNewID(t *testing.T) {
	id := NewID("test-")
	if len(id) != 13 {
		t.Errorf("NewID() = %s; want length 13", id)
	}
	if id[:5] != "test-" {
		t.Errorf("NewID() = %s; want prefix 'test-'", id)
	}
}
