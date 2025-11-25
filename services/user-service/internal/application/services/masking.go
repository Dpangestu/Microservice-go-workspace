package services

import (
	"log"
	"strings"
)

// Scope constants
const (
	ScopeProfileEmail     = "profile email"
	ScopeProfileBasic     = "profile"
	ScopeProfileSensitive = "profile:sensitive"
)

// MaskUserBundleByScope menghapus field sensitif berdasarkan scope JWT
func MaskUserBundleByScope(bundle map[string]interface{}, scope string) map[string]interface{} {
	scopes := parseScopes(scope)

	// Default: mask semua field sensitif
	if !hasScope(scopes, "profile:sensitive") {
		// Hapus field sensitif dari data
		if data, ok := bundle["data"].(map[string]interface{}); ok {
			// Bisa mask failedLoginAttempts, isLocked untuk non-sensitive scope
			if !hasScope(scopes, "profile:admin") {
				delete(data, "failedLoginAttempts")
				delete(data, "isLocked")
			}
		}

		// Hapus field sensitif dari profile metadata jika ada
		if profile, ok := bundle["profile"].(map[string]interface{}); ok {
			if metadata, ok := profile["metadata"].(map[string]interface{}); ok {
				// Hanya tampilkan metadata public fields
				publicMetadata := make(map[string]interface{})
				for k, v := range metadata {
					// Whitelist: hanya certain fields yang boleh di-expose
					if k == "title" {
						publicMetadata[k] = v
					}
				}
				if len(publicMetadata) > 0 {
					profile["metadata"] = publicMetadata
				} else {
					delete(profile, "metadata")
				}
			}
		}
	}

	// Hapus settings kecuali user request dengan scope khusus
	if !hasScope(scopes, "profile:settings") && !hasScope(scopes, "profile:sensitive") {
		delete(bundle, "settings")
	}

	log.Printf("[Masking] Applied masking for scopes: %v", scopes)
	return bundle
}

// parseScopes convert "scope1 scope2" string ke slice
func parseScopes(scope string) []string {
	return strings.Fields(strings.TrimSpace(scope))
}

// hasScope check apakah scope list mengandung scope tertentu
func hasScope(scopes []string, target string) bool {
	for _, s := range scopes {
		if s == target {
			return true
		}
	}
	return false
}
