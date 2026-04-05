package service

import (
	"fmt"
	"strings"
)

type ScopeDefinition struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Description string `json:"description"`
}

var scopeCatalog = []ScopeDefinition{
	{Key: "openid", Label: "OpenID", Description: "Authenticate the user and issue an ID token."},
	{Key: "profile", Label: "Profile", Description: "Read basic account profile."},
	{Key: "email", Label: "Email", Description: "Read account email."},
	{Key: "roles", Label: "Roles", Description: "Read role information."},
	{Key: "offline_access", Label: "Offline Access", Description: "Issue refresh tokens for long-lived sessions."},
}

func DefaultScopes() []string {
	return []string{"profile"}
}

func ScopeDefinitions() []ScopeDefinition {
	items := make([]ScopeDefinition, len(scopeCatalog))
	copy(items, scopeCatalog)
	return items
}

func NormalizeKnownScopes(scopes []string) ([]string, error) {
	normalized := normalizeScopeList(scopes)
	if len(normalized) == 0 {
		return DefaultScopes(), nil
	}
	for _, scope := range normalized {
		if _, ok := scopeDefinition(scope); !ok {
			return nil, fmt.Errorf("%w: unsupported scope %s", ErrInvalidInput, scope)
		}
	}
	return normalized, nil
}

func ParseKnownScopeString(scope string) []string {
	items := normalizeScopeList(strings.Fields(strings.ReplaceAll(strings.TrimSpace(scope), ",", " ")))
	if len(items) == 0 {
		return nil
	}
	knownScopes := make([]string, 0, len(items))
	for _, item := range items {
		if _, ok := scopeDefinition(item); ok {
			knownScopes = append(knownScopes, item)
		}
	}
	return knownScopes
}

func HasScope(scopeValue string, required string) bool {
	granted := ParseKnownScopeString(scopeValue)
	for _, scope := range granted {
		if scope == required {
			return true
		}
	}
	return false
}

func ScopeDetails(scopes []string) []map[string]any {
	normalized, err := NormalizeKnownScopes(scopes)
	if err != nil {
		normalized = DefaultScopes()
	}
	items := make([]map[string]any, 0, len(normalized))
	for _, scope := range normalized {
		definition, _ := scopeDefinition(scope)
		items = append(items, map[string]any{
			"key":         definition.Key,
			"label":       definition.Label,
			"description": definition.Description,
		})
	}
	return items
}

func scopeDefinition(scope string) (ScopeDefinition, bool) {
	for _, definition := range scopeCatalog {
		if definition.Key == scope {
			return definition, true
		}
	}
	return ScopeDefinition{}, false
}
