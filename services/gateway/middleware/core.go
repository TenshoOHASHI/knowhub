package middleware

import "net/http"

type CoreMiddleware struct {
	AllowedOrigin     string
	AllowedMethods    string
	AllowedHeaders    string
	AllowedCredential string
}

func NewCoreMiddleware(origin string, methods string, headers string, credential string) *CoreMiddleware {
	return &CoreMiddleware{AllowedOrigin: origin, AllowedMethods: methods, AllowedHeaders: headers, AllowedCredential: credential}
}

func (c CoreMiddleware) CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", c.AllowedOrigin)
		w.Header().Set("Access-Control-Allow-Methods", c.AllowedMethods)
		w.Header().Set("Access-Control-Allow-Headers", c.AllowedHeaders)
		w.Header().Set("Access-Control-Allow-Credentials", c.AllowedCredential)

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
