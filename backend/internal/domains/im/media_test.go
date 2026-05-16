package im

import "testing"

func TestParseMediaRef(t *testing.T) {
	t.Parallel()

	ref := BuildInboundMediaRef("feishu", "image", "img_key_1")
	platform, attachmentType, remoteID, ok := ParseMediaRef(ref)
	if !ok {
		t.Fatal("ParseMediaRef() ok = false, want true")
	}
	if platform != "feishu" {
		t.Fatalf("platform = %q, want feishu", platform)
	}
	if attachmentType != "image" {
		t.Fatalf("attachmentType = %q, want image", attachmentType)
	}
	if remoteID != "img_key_1" {
		t.Fatalf("remoteID = %q, want img_key_1", remoteID)
	}
}

func TestParseMediaRefInvalid(t *testing.T) {
	t.Parallel()

	tests := []string{
		"",
		"media://",
		"media://feishu/image",
		"media://feishu/image/not-base64!",
		"https://example.com/file",
	}
	for _, ref := range tests {
		ref := ref
		t.Run(ref, func(t *testing.T) {
			t.Parallel()

			if _, _, _, ok := ParseMediaRef(ref); ok {
				t.Fatalf("ParseMediaRef(%q) ok = true, want false", ref)
			}
		})
	}
}
