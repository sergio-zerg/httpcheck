package result

import (
	"strings"
	"testing"
	"time"
)

func TestNewSetsFields(t *testing.T) {
	r := New("1.2.3.4", "check:http://x/", "All OK", 250*time.Millisecond, false)
	if r.IP != "1.2.3.4" || r.Name != "check:http://x/" || r.Message != "All OK" {
		t.Fatalf("unexpected fields: %+v", r)
	}
	if r.IsError {
		t.Fatalf("expected IsError=false")
	}
	if r.Duration != 250*time.Millisecond {
		t.Fatalf("unexpected duration: %v", r.Duration)
	}
	if r.Date.IsZero() {
		t.Fatalf("Date should be stamped")
	}
}

func TestStringContainsParts(t *testing.T) {
	r := New("ip", "my-check", "boom", 1200*time.Millisecond, true)
	s := r.String()
	for _, part := range []string{"my-check", "boom", "1.2s"} {
		if !strings.Contains(s, part) {
			t.Errorf("String()=%q missing %q", s, part)
		}
	}
}
