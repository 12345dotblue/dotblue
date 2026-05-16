package im

import "testing"

func TestParseFeishuContentPost(t *testing.T) {
	t.Parallel()

	content := `{
		"post": {
			"zh_cn": {
				"title": "demo",
				"content": [
					[
						{"tag":"text","text":"hello "},
						{"tag":"a","text":"dotblue","href":"https://dotblue.ai"},
						{"tag":"at","user_id":"ou_xxx","user_name":"bot"},
						{"tag":"img","image_key":"img_xxx","alt":"screenshot"}
					]
				]
			}
		}
	}`

	result := parseFeishuContent("post", content, "feishu")
	if result.Text != "hello dotbluebot" {
		t.Fatalf("Text = %q, want flattened post text", result.Text)
	}
	if len(result.Segments) < 4 {
		t.Fatalf("Segments len = %d, want >= 4", len(result.Segments))
	}
	if len(result.Attachments) != 1 {
		t.Fatalf("Attachments len = %d, want 1", len(result.Attachments))
	}
	if result.Attachments[0].MediaRef == "" {
		t.Fatal("attachment media ref is empty")
	}
}

func TestParseFeishuContentImage(t *testing.T) {
	t.Parallel()

	result := parseFeishuContent("image", `{"image_key":"img_xxx"}`, "feishu")
	if len(result.Attachments) != 1 {
		t.Fatalf("Attachments len = %d, want 1", len(result.Attachments))
	}
	if result.Attachments[0].Type != "image" {
		t.Fatalf("Attachment type = %q, want image", result.Attachments[0].Type)
	}
	if got := result.Attachments[0].MediaRef; got == "" {
		t.Fatal("MediaRef is empty")
	}
}
