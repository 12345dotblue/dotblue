package metering

import "testing"

func TestNullableUUIDValue(t *testing.T) {
	if got := nullableUUIDValue(""); got != nil {
		t.Fatalf("nullableUUIDValue(\"\") = %#v, want nil", got)
	}
	if got := nullableUUIDValue("  "); got != nil {
		t.Fatalf("nullableUUIDValue(\"  \") = %#v, want nil", got)
	}
	if got := nullableUUIDValue(" abc "); got != "abc" {
		t.Fatalf("nullableUUIDValue(\" abc \") = %#v, want %q", got, "abc")
	}
}

func TestNormalizedJSONBValue(t *testing.T) {
	if got := normalizedJSONBValue(""); got != "{}" {
		t.Fatalf("normalizedJSONBValue(\"\") = %q, want {}", got)
	}
	if got := normalizedJSONBValue("  "); got != "{}" {
		t.Fatalf("normalizedJSONBValue(\"  \") = %q, want {}", got)
	}
	if got := normalizedJSONBValue(`{"ok":true}`); got != `{"ok":true}` {
		t.Fatalf("normalizedJSONBValue(valid) = %q, want unchanged", got)
	}
}
