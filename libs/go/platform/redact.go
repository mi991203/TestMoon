package platform

import (
	"strings"
)

func Sensitive(k string) bool {
	k = strings.ToLower(strings.ReplaceAll(k, "-", "_"))
	return strings.Contains(k, "password") || strings.Contains(k, "token") || strings.Contains(k, "secret") || k == "authorization" || k == "cookie" || k == "set_cookie" || k == "x_api_key" || k == "api_key" || k == "apikey" || strings.Contains(k, "access_key")
}
