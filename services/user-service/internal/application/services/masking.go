package services

import (
	"log"
	"strings"
)

// MaskUserBundleByScope menghapus field sensitif berdasarkan JWT scope
func MaskUserBundleByScope(bundle map[string]interface{}, scope string) map[string]interface{} {
	scopes := parseScopes(scope)

	// Default: mask semua field sensitif
	if !hasScope(scopes, "profile:sensitive") && !hasScope(scopes, "profile:admin") {
		// Hapus field sensitif dari data
		if data, ok := bundle["data"].(map[string]interface{}); ok {
			// Hide sensitive fields untuk non-admin scope
			delete(data, "failedLoginAttempts")
			delete(data, "isLocked")
		}

		// Mask metadata sensitif dari profile
		if profile, ok := bundle["profile"].(map[string]interface{}); ok {
			if metadata, ok := profile["metadata"].(map[string]interface{}); ok {
				// Hanya expose non-sensitive metadata fields
				publicMetadata := make(map[string]interface{})
				whitelistedKeys := map[string]bool{
					"title":      true,
					"department": true,
				}

				for k, v := range metadata {
					if whitelistedKeys[k] {
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

	// Hapus settings untuk non-authorized scope
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

// hasScope check apakah scope list mengandung target
func hasScope(scopes []string, target string) bool {
	for _, s := range scopes {
		if s == target {
			return true
		}
	}
	return false
}
