package im

import "strings"

type feishuContentParseResult struct {
	Text        string
	Segments    []RichSegment
	Attachments []InboundAttachment
}

func parseFeishuContent(messageType, content, platform string) feishuContentParseResult {
	messageType = strings.TrimSpace(messageType)
	content = strings.TrimSpace(content)
	if content == "" {
		return feishuContentParseResult{}
	}

	payload, err := coerceRawMap(content)
	if err != nil {
		text := strings.TrimSpace(content)
		if text == "" {
			return feishuContentParseResult{}
		}
		return feishuContentParseResult{
			Text:     text,
			Segments: []RichSegment{{Type: "text", Text: text}},
		}
	}

	if post, ok := payload["post"].(map[string]any); ok {
		return parseFeishuPostContent(post, platform)
	}

	switch messageType {
	case "image":
		return parseFeishuStandaloneAttachment(payload, platform, "image", "image_key", "image")
	case "file":
		return parseFeishuStandaloneAttachment(payload, platform, "file", "file_key", "file_name")
	case "audio":
		return parseFeishuStandaloneAttachment(payload, platform, "audio", "file_key", "file_name")
	case "media":
		return parseFeishuStandaloneAttachment(payload, platform, "video", "file_key", "file_name")
	case "sticker":
		return parseFeishuStandaloneAttachment(payload, platform, "image", "file_key", "file_name")
	}

	if text := str(payload["text"]); text != "" {
		return feishuContentParseResult{
			Text:     text,
			Segments: []RichSegment{{Type: "text", Text: text}},
		}
	}

	return feishuContentParseResult{
		Text: strings.TrimSpace(content),
	}
}

func parseFeishuStandaloneAttachment(
	payload map[string]any,
	platform string,
	attachmentType string,
	keyField string,
	nameField string,
) feishuContentParseResult {
	remoteID := str(payload[keyField])
	name := str(payload[nameField])
	attachment := buildInboundAttachment(platform, attachmentType, remoteID, name)
	if attachment.MediaRef == "" {
		return feishuContentParseResult{}
	}
	return feishuContentParseResult{
		Segments: []RichSegment{{
			Type: attachmentType,
			Meta: map[string]any{
				"media_ref": attachment.MediaRef,
				"name":      attachment.Name,
			},
		}},
		Attachments: []InboundAttachment{attachment},
	}
}

func parseFeishuPostContent(post map[string]any, platform string) feishuContentParseResult {
	localeMap := pickFirstMap(post)
	if localeMap == nil {
		return feishuContentParseResult{}
	}
	rows, _ := localeMap["content"].([]any)

	result := feishuContentParseResult{}
	for _, row := range rows {
		elements, _ := row.([]any)
		for _, rawElement := range elements {
			element, _ := rawElement.(map[string]any)
			tag := str(element["tag"])
			switch tag {
			case "text":
				text := rawString(element["text"])
				if text != "" {
					result.Segments = append(result.Segments, RichSegment{Type: "text", Text: text})
				}
			case "a":
				text := firstNonEmpty(rawString(element["text"]), str(element["href"]))
				if text != "" {
					result.Segments = append(result.Segments, RichSegment{
						Type: "link",
						Text: text,
						Meta: map[string]any{
							"url": str(element["href"]),
						},
					})
				}
			case "at":
				text := firstNonEmpty(rawString(element["user_name"]), rawString(element["text"]), "@"+str(element["user_id"]))
				if text != "" {
					result.Segments = append(result.Segments, RichSegment{
						Type: "mention",
						Text: text,
						Meta: map[string]any{
							"user_id": str(element["user_id"]),
						},
					})
				}
			case "img":
				attachment := buildInboundAttachment(platform, "image", str(element["image_key"]), str(element["alt"]))
				appendAttachmentSegment(&result, attachment, "image")
			case "media":
				attachmentType := firstNonEmpty(str(element["media_type"]), "file")
				attachment := buildInboundAttachment(platform, attachmentType, str(element["file_key"]), str(element["file_name"]))
				appendAttachmentSegment(&result, attachment, attachmentType)
			case "file":
				attachment := buildInboundAttachment(platform, "file", str(element["file_key"]), str(element["file_name"]))
				appendAttachmentSegment(&result, attachment, "file")
			case "emotion":
				text := firstNonEmpty(str(element["text"]), ":"+str(element["emoji_type"])+":")
				if text != "" {
					result.Segments = append(result.Segments, RichSegment{Type: "text", Text: text})
				}
			}
		}
		if len(elements) > 0 {
			result.Segments = append(result.Segments, RichSegment{Type: "text", Text: "\n"})
		}
	}

	result.Text = strings.TrimSpace(segmentsPlainText(result.Segments))
	return result
}

func rawString(v any) string {
	s, _ := v.(string)
	return s
}

func appendAttachmentSegment(result *feishuContentParseResult, attachment InboundAttachment, segmentType string) {
	if attachment.MediaRef == "" {
		return
	}
	result.Attachments = append(result.Attachments, attachment)
	result.Segments = append(result.Segments, RichSegment{
		Type: segmentType,
		Meta: map[string]any{
			"media_ref": attachment.MediaRef,
			"name":      attachment.Name,
		},
	})
}

func pickFirstMap(in map[string]any) map[string]any {
	for _, value := range in {
		if out, ok := value.(map[string]any); ok {
			return out
		}
	}
	return nil
}

func segmentsPlainText(segments []RichSegment) string {
	var builder strings.Builder
	for _, segment := range segments {
		switch segment.Type {
		case "text", "link", "mention", "quote", "code":
			builder.WriteString(segment.Text)
		}
	}
	return builder.String()
}
