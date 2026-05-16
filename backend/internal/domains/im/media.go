package im

import (
	"encoding/base64"
	"strings"
)

func BuildInboundMediaRef(platform, attachmentType, remoteID string) string {
	platform = strings.TrimSpace(platform)
	attachmentType = strings.TrimSpace(attachmentType)
	remoteID = strings.TrimSpace(remoteID)
	if platform == "" || attachmentType == "" || remoteID == "" {
		return ""
	}
	token := base64.RawURLEncoding.EncodeToString([]byte(remoteID))
	return "media://" + platform + "/" + attachmentType + "/" + token
}

func ParseMediaRef(ref string) (platform string, attachmentType string, remoteID string, ok bool) {
	ref = strings.TrimSpace(ref)
	if !strings.HasPrefix(ref, "media://") {
		return "", "", "", false
	}
	parts := strings.Split(strings.TrimPrefix(ref, "media://"), "/")
	if len(parts) != 3 {
		return "", "", "", false
	}
	decoded, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(parts[2]))
	if err != nil {
		return "", "", "", false
	}
	platform = strings.TrimSpace(parts[0])
	attachmentType = strings.TrimSpace(parts[1])
	remoteID = strings.TrimSpace(string(decoded))
	if platform == "" || attachmentType == "" || remoteID == "" {
		return "", "", "", false
	}
	return platform, attachmentType, remoteID, true
}

func detectAttachmentContentType(attachmentType string) string {
	switch strings.TrimSpace(attachmentType) {
	case "image":
		return "image/*"
	case "audio":
		return "audio/*"
	case "video":
		return "video/*"
	case "file":
		return "application/octet-stream"
	default:
		return ""
	}
}

func buildInboundAttachment(platform, attachmentType, remoteID, name string) InboundAttachment {
	attachmentType = strings.TrimSpace(attachmentType)
	return InboundAttachment{
		Type:        attachmentType,
		Name:        strings.TrimSpace(name),
		ContentType: detectAttachmentContentType(attachmentType),
		MediaRef:    BuildInboundMediaRef(platform, attachmentType, remoteID),
	}
}
