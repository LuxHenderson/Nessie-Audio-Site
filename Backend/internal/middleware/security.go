package middleware

import (
	"net/http"
	"strings"
)

// SecurityHeaders adds security headers to all responses
func SecurityHeaders() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Strict-Transport-Security (HSTS)
			// Tells browsers to ONLY use HTTPS for this domain for 1 year
			// includeSubDomains: Apply to all subdomains
			// preload: Allow inclusion in browser HSTS preload lists
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")

			// X-Content-Type-Options
			// Prevents browsers from MIME-sniffing away from declared content-type
			// Stops attackers from disguising malicious files as safe types
			w.Header().Set("X-Content-Type-Options", "nosniff")

			// X-Frame-Options
			// Prevents page from being loaded in iframe (clickjacking protection)
			// SAMEORIGIN: Allow framing only by same origin
			w.Header().Set("X-Frame-Options", "SAMEORIGIN")

			// X-XSS-Protection
			// Legacy XSS filter for older browsers (IE, Chrome, Safari)
			// Modern browsers rely on CSP instead
			// 1; mode=block: Enable XSS filter and block page if attack detected
			w.Header().Set("X-XSS-Protection", "1; mode=block")

			// Content-Security-Policy (CSP)
			// Controls what resources the browser can load
			// This is a balanced policy for an eCommerce site
			csp := strings.Join([]string{
				"default-src 'self'",                          // Only load from same origin by default
				"script-src 'self' 'unsafe-inline' https://js.stripe.com https://cdn.jsdelivr.net https://static.cloudflareinsights.com", // Allow scripts from self, Stripe, Three.js CDN, Cloudflare analytics
				"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com https://cdnjs.cloudflare.com", // Allow inline styles, Google Fonts, Font Awesome
				"img-src 'self' data: https:",                 // Allow images from self, data URIs, and HTTPS
				"font-src 'self' data: https://fonts.gstatic.com https://cdnjs.cloudflare.com", // Allow fonts from self, data URIs, Google Fonts, Font Awesome
				"connect-src 'self' https://api.stripe.com https://api.printful.com", // Allow API calls to self, Stripe, Printful
				"frame-src https://js.stripe.com",             // Allow Stripe iframe for payment
				"object-src 'none'",                           // Block Flash, Java, etc.
				"base-uri 'self'",                             // Restrict <base> tag
				"form-action 'self' https://formspree.io",     // Allow form submissions to same origin and Formspree
				"frame-ancestors 'self'",                      // Only allow framing by same origin
				"upgrade-insecure-requests",                   // Upgrade HTTP requests to HTTPS
			}, "; ")
			w.Header().Set("Content-Security-Policy", csp)

			// Referrer-Policy
			// Controls how much referrer information is sent with requests
			// strict-origin-when-cross-origin: Send full URL for same-origin, only origin for cross-origin HTTPS
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

			// Permissions-Policy (formerly Feature-Policy)
			// Controls browser features and APIs
			// Disable microphone, camera, geolocation, payment (we use Stripe instead)
			permissionsPolicy := strings.Join([]string{
				"microphone=()",
				"camera=()",
				"geolocation=()",
				"interest-cohort=()", // Disable FLoC (privacy)
			}, ", ")
			w.Header().Set("Permissions-Policy", permissionsPolicy)

			next.ServeHTTP(w, r)
		})
	}
}

// HTTPSRedirect redirects HTTP requests to HTTPS in production
// In development, it does nothing (allows localhost HTTP)
func HTTPSRedirect(env string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip redirect in development
			if env == "development" {
				next.ServeHTTP(w, r)
				return
			}

			// Check if request is already HTTPS
			isHTTPS := r.TLS != nil ||
				r.Header.Get("X-Forwarded-Proto") == "https" ||
				r.URL.Scheme == "https"

			if !isHTTPS {
				// Build HTTPS URL
				httpsURL := "https://" + r.Host + r.RequestURI

				// Permanent redirect (301)
				http.Redirect(w, r, httpsURL, http.StatusMovedPermanently)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
