package im

import (
	"crypto/sha1"
	"encoding/hex"
	"strings"
)

const (
	SessionStrategyPerUser        = "per_user"
	SessionStrategyPerChat        = "per_chat"
	SessionStrategyPerThread      = "per_thread"
	SessionStrategyPerChatPerUser = "per_chat_per_user"
)

func BuildSessionAddress(conn Connection, event InboundEvent) SessionAddress {
	return SessionAddress{
		Platform:     strings.TrimSpace(firstNonEmpty(event.Platform, conn.Platform)),
		EnterpriseID: strings.TrimSpace(firstNonEmpty(event.EnterpriseID, conn.EnterpriseID)),
		ConnectionID: strings.TrimSpace(firstNonEmpty(event.ConnectionID, conn.ID)),
		ChatID:       strings.TrimSpace(event.ExternalChatID),
		ThreadID:     strings.TrimSpace(event.ExternalThreadID),
		UserID:       strings.TrimSpace(event.ExternalUserID),
		ChatType:     normalizeChatType(event.ChatType),
	}
}

func BuildSessionKey(agentID, strategy string, addr SessionAddress) string {
	parts := []string{
		"agent",
		sanitizeSessionPart(agentID),
		sanitizeSessionPart(addr.Platform),
		sanitizeSessionPart(addr.ConnectionID),
	}

	switch normalizeSessionStrategy(strategy, addr) {
	case SessionStrategyPerUser:
		parts = append(parts, "user", sanitizeSessionPart(addr.UserID))
	case SessionStrategyPerThread:
		parts = append(parts, "thread", sanitizeSessionPart(firstNonEmpty(addr.ThreadID, addr.ChatID, addr.UserID)))
	case SessionStrategyPerChatPerUser:
		parts = append(parts, "chat_user", sanitizeSessionPart(addr.ChatID), sanitizeSessionPart(addr.UserID))
	default:
		parts = append(parts, "chat", sanitizeSessionPart(firstNonEmpty(addr.ChatID, addr.UserID)))
	}

	return strings.Join(parts, ":")
}

func normalizeSessionStrategy(strategy string, addr SessionAddress) string {
	switch strings.TrimSpace(strategy) {
	case SessionStrategyPerUser, SessionStrategyPerChat, SessionStrategyPerThread, SessionStrategyPerChatPerUser:
		return strings.TrimSpace(strategy)
	}
	if strings.TrimSpace(addr.ThreadID) != "" {
		return SessionStrategyPerThread
	}
	switch normalizeChatType(addr.ChatType) {
	case "p2p", "dm", "direct":
		return SessionStrategyPerUser
	default:
		return SessionStrategyPerChatPerUser
	}
}

func sanitizeSessionPart(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "_"
	}
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"\n", "_",
		"\r", "_",
		"\t", "_",
		" ", "_",
	)
	value = replacer.Replace(value)
	for strings.Contains(value, "__") {
		value = strings.ReplaceAll(value, "__", "_")
	}
	value = strings.Trim(value, "_")
	if value == "" {
		return "_"
	}
	if len(value) > 96 {
		sum := sha1.Sum([]byte(value))
		return value[:48] + "_" + hex.EncodeToString(sum[:6])
	}
	return value
}

func normalizeChatType(chatType string) string {
	chatType = strings.TrimSpace(strings.ToLower(chatType))
	switch chatType {
	case "p2p", "dm", "direct":
		return "p2p"
	case "group", "channel":
		return "group"
	default:
		return chatType
	}
}
