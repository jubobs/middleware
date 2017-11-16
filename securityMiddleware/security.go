// Package securityMiddleware adds various security-related headers.
package securityMiddleware // import "github.com/teamwork/middleware/securityMiddleware"

import (
	"fmt"
	"net/http"
	"strings"
)

// Config defines the config for Security middleware.
type Config struct {
	// XFrameOptions controls where this site can be displayed in a frame.
	//
	// DENY
	//     The page cannot be displayed in a frame, regardless of the site
	//     attempting to do so.
	// SAMEORIGIN
	//     The page can only be displayed in a frame on the same origin as the
	//     page itself.
	// ALLOW-FROM uri
	//     The page can only be displayed in a frame on the specified origin.
	//
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Frame-Options
	XFrameOptions string

	// ContentSecurityPolicy controls which JS and CSS resources can be run,
	// preventing XSS attacks
	//
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Security-Policy
	ContentSecurityPolicy map[string][]string

	// ContentSecurityPolicyReportOnly is like the CSP header, but only reports
	// violations and doesn't block anything.
	ContentSecurityPolicyReportOnly map[string][]string

	// StrictTransportSecurity makes sure that browsers only communicate over
	// https.
	//
	// Note: right now this only affects *.teamwork.com domains and not custom
	// domains!
	//
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Strict-Transport-Security
	StrictTransportSecurity string

	// XContentTypeOptions makes sure that browsers don't autoguess the
	// Content-Type, preventing certain attacks.
	//
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Content-Type-Options
	XContentTypeOptions string
}

// DefaultConfig is the default Security middleware config.
var DefaultConfig = Config{
	XFrameOptions: "SAMEORIGIN",

	// 30 days
	StrictTransportSecurity: "max-age=2592000",

	// JS and CSS must be from our own domain, but images can be from anywhere.
	// We also use inline scripts in a lot of places, so unfortunatly we'll have
	// to enable that for now as well :-(
	// We only report for now, since this is new!
	/*
		ContentSecurityPolicyReportOnly: map[string][]string{
			"default-src": {"'self'"},
			"report-uri":  {"/desk/csp"},

			"font-src": {
				"'self'",
				// Font-"Awesome"
				"https://cdnjs.cloudflare.com",
				"https://maxcdn.bootstrapcdn.com",
			},

			// Need to allow loading images from emails etc.
			"img-src":    {"*", "data:", "'unsafe-inline'"},
			"script-src": {"*", "'unsafe-eval'", "'unsafe-inline'"},
			"style-src":  {"*", "'unsafe-inline'"},

			"connect-src": {
				"'self'",

				// TODO: Get installation code for this domain
				// sec.ContentSecurityPolicyReportOnly["connect-src"] = append(
				// 	sec.ContentSecurityPolicyReportOnly["connect-src"],
				// 	fmt.Sprintf("ws://%v%v", "", viper.GetString("websocket.bind")))
			},

			"child-src": {
			},

			"media-src": {
			},
		},
	*/

	// Blocks a request if the requested type is:
	// - "style" and the MIME type is not "text/css"
	// - "script" and the MIME type is not a JavaScript MIME type.
	XContentTypeOptions: "nosniff",
}

// Add sets several security-related headers.
func Add(rootDomain string) func(http.Handler) http.Handler {
	return WithConfig(DefaultConfig, rootDomain)
}

// WithConfig returns a Security middleware from config.
func WithConfig(config Config, rootDomain string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			csp := ""
			for k, v := range config.ContentSecurityPolicy {
				csp += fmt.Sprintf("%v %v;", k, strings.Join(v, " "))
			}

			cspReport := ""
			for k, v := range config.ContentSecurityPolicyReportOnly {
				cspReport += fmt.Sprintf("%v %v;", k, strings.Join(v, " "))
			}

			if config.XFrameOptions != "" {
				w.Header().Set("X-Frame-Options", config.XFrameOptions)
			}
			if csp != "" {
				w.Header().Set("Content-Security-Policy", csp)
			}
			if cspReport != "" {
				w.Header().Set("Content-Security-Policy-Report-Only", cspReport)
			}
			if config.StrictTransportSecurity != "" &&
				strings.HasSuffix(r.Host, rootDomain) {
				w.Header().Set("Strict-Transport-Security", config.StrictTransportSecurity)
			}
			if config.XContentTypeOptions != "" {
				w.Header().Set("X-Content-Type-Options", config.XContentTypeOptions)
			}

			next.ServeHTTP(w, r)
		})
	}
}
