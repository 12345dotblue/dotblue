package im

import (
	"encoding/json"
	"strings"
)

type feishuOutboundMessage struct {
	MsgType string
	Content string
}

func buildFeishuOutboundMessage(msg OutboundEnvelope) (feishuOutboundMessage, error) {
	if shouldUseFeishuPost(msg) {
		content, err := buildFeishuPostContent(msg)
		if err == nil && content != "" {
			return feishuOutboundMessage{
				MsgType: "post",
				Content: content,
			}, nil
		}
	}

	text := formatFeishuOutboundText(msg)
	content, err := json.Marshal(map[string]string{"text": text})
	if err != nil {
		return feishuOutboundMessage{}, err
	}
	return feishuOutboundMessage{
		MsgType: "text",
		Content: string(content),
	}, nil
}

func shouldUseFeishuPost(msg OutboundEnvelope) bool {
	if len(msg.Attachments) > 0 || len(msg.Segments) == 0 {
		return false
	}
	for _, segment := range msg.Segments {
		switch segment.Type {
		case "link", "mention":
			return true
		}
	}
	return false
}

func buildFeishuPostContent(msg OutboundEnvelope) (string, error) {
	rows := buildFeishuPostRows(msg.Segments)
	if len(rows) == 0 {
		return "", nil
	}
	payload := map[string]any{
		"post": map[string]any{
			"zh_cn": map[string]any{
				"title":   "",
				"content": rows,
			},
		},
	}
	content, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func buildFeishuPostRows(segments []RichSegment) []any {
	rows := make([]any, 0)
	current := make([]any, 0)

	flush := func() {
		if len(current) == 0 {
			return
		}
		row := make([]any, len(current))
		copy(row, current)
		rows = append(rows, row)
		current = make([]any, 0)
	}

	for _, segment := range segments {
		switch segment.Type {
		case "text", "code", "quote":
			parts := strings.Split(segment.Text, "\n")
			for i, part := range parts {
				if part != "" {
					current = append(current, map[string]any{
						"tag":  "text",
						"text": part,
					})
				}
				if i < len(parts)-1 {
					flush()
				}
			}
		case "link":
			current = append(current, map[string]any{
				"tag":  "a",
				"text": firstNonEmpty(segment.Text, str(segment.Meta["url"])),
				"href": str(segment.Meta["url"]),
			})
		case "mention":
			current = append(current, map[string]any{
				"tag":       "at",
				"user_id":   firstNonEmpty(str(segment.Meta["user_id"]), str(segment.Meta["open_id"])),
				"user_name": segment.Text,
			})
		default:
			if segment.Text != "" {
				current = append(current, map[string]any{
					"tag":  "text",
					"text": segment.Text,
				})
			}
		}
	}
	flush()
	return rows
}
