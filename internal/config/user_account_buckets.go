package config

import (
	"fmt"
	"strings"
)

const defaultProfileKey = "default"

// UserAccountBucketKey returns the bucket key used to map (app_id, base_url, profile)
// to a user account label.
func UserAccountBucketKey(appID, baseURL, profile string) string {
	appID = strings.TrimSpace(appID)
	baseURL = normalizeBucketBaseURL(baseURL)
	profile = strings.TrimSpace(profile)
	if profile == "" || strings.EqualFold(profile, defaultProfileKey) {
		profile = defaultProfileKey
	}
	return fmt.Sprintf("%s|%s|%s", appID, baseURL, profile)
}

func normalizeBucketBaseURL(baseURL string) string {
	base := strings.TrimSpace(baseURL)
	base = strings.TrimRight(base, "/")
	base = strings.TrimSuffix(base, "/open-apis")
	base = strings.TrimSuffix(base, "/open-apis/")
	base = strings.TrimRight(base, "/")
	return strings.ToLower(base)
}
