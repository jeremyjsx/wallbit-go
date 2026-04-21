package wallbit

import (
	"testing"
	"time"
)

func TestPtrReturnsAddressOfCopy(t *testing.T) {
	t.Parallel()

	src := "jeremy"
	p := Ptr(src)
	if p == nil || *p != "jeremy" {
		t.Fatalf("Ptr(%q) dereferenced to %v", src, p)
	}
	// Mutating the returned pointee must not leak back into the
	// caller's variable: Ptr takes its argument by value.
	*p = "mutated"
	if src != "jeremy" {
		t.Fatalf("Ptr aliased caller's variable: src = %q after mutation", src)
	}
}

func TestPtrPreservesTypeForNumericAndStruct(t *testing.T) {
	t.Parallel()

	if got := *Ptr(42); got != 42 {
		t.Fatalf("Ptr(int) = %d, want 42", got)
	}
	if got := *Ptr(3.14); got != 3.14 {
		t.Fatalf("Ptr(float64) = %v, want 3.14", got)
	}

	type address struct {
		Street string
		City   string
		Zip    string
	}
	a := address{Street: "Av. Corrientes", City: "Buenos Aires", Zip: "C1043"}
	p := Ptr(a)
	if p == nil || *p != a {
		t.Fatalf("Ptr(struct) = %+v, want %+v", p, a)
	}
}

func TestPtrWithZeroValues(t *testing.T) {
	t.Parallel()

	empty := Ptr("")
	if empty == nil || *empty != "" {
		t.Fatalf("Ptr(\"\") returned %v, want non-nil pointer to empty string", empty)
	}
	zero := Ptr(time.Duration(0))
	if zero == nil || *zero != 0 {
		t.Fatalf("Ptr(0) returned %v, want non-nil pointer to zero duration", zero)
	}
}
