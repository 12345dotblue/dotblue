package chatentry

import (
	"context"
	"testing"
)

func TestJWTTokenIssuerIssueAndParseSessionToken(t *testing.T) {
	issuer := &jwtTokenIssuer{secret: []byte("test-secret")}
	token, _, err := issuer.IssueSessionToken(context.Background(), SessionTokenClaims{
		EnterpriseID: "ent-1",
		AgentID:      "agent-1",
		ProxyUserID:  "user-1",
		SubjectType:  "share_visitor",
		SubjectID:    "share-1",
		Scopes:       []string{ScopeChatRead, ScopeChatWrite},
		AccessMode:   AccessModeShare,
	})
	if err != nil {
		t.Fatalf("issue session token: %v", err)
	}

	claims, err := issuer.ParseSessionToken(context.Background(), token)
	if err != nil {
		t.Fatalf("parse session token: %v", err)
	}
	if claims.AgentID != "agent-1" {
		t.Fatalf("expected agent-1, got %s", claims.AgentID)
	}
	if claims.ProxyUserID != "user-1" {
		t.Fatalf("expected user-1, got %s", claims.ProxyUserID)
	}
	if len(claims.Scopes) != 2 {
		t.Fatalf("expected two scopes, got %d", len(claims.Scopes))
	}
}

func TestJWTTokenIssuerIssueAndParseEmbedToken(t *testing.T) {
	issuer := &jwtTokenIssuer{secret: []byte("test-secret")}
	token, _, err := issuer.IssueEmbedToken(context.Background(), EmbedTokenClaims{
		EnterpriseID:  "ent-1",
		AgentID:       "agent-1",
		ProxyUserID:   "user-1",
		Origin:        "https://example.com",
		EmbedConfigID: "embed-1",
	})
	if err != nil {
		t.Fatalf("issue embed token: %v", err)
	}

	claims, err := issuer.ParseEmbedToken(context.Background(), token)
	if err != nil {
		t.Fatalf("parse embed token: %v", err)
	}
	if claims.Origin != "https://example.com" {
		t.Fatalf("expected origin https://example.com, got %s", claims.Origin)
	}
	if claims.EmbedConfigID != "embed-1" {
		t.Fatalf("expected embed-1, got %s", claims.EmbedConfigID)
	}
}
