package engine

import (
	"testing"
)

func Test_containerName(t *testing.T) {
	name := containerName("550e8400-e29b-41d4-a716-446655440000")
	expected := "hermes_550e8400-e29b-41d4-a716-446655440000"
	if name != expected {
		t.Errorf("containerName() = %q, want %q", name, expected)
	}
}

func Test_profileVolumePath(t *testing.T) {
	path := profileVolumePath("/data/hermes", "550e8400-e29b-41d4-a716-446655440000")
	expected := "/data/hermes/550e8400-e29b-41d4-a716-446655440000"
	if path != expected {
		t.Errorf("profileVolumePath() = %q, want %q", path, expected)
	}
}
