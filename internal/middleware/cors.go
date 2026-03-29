package middleware

import (
	"net/http"
	"strings"

	"github.com/go-chi/cors"
)

func CORS(allowedOriginsStr string) func(http.Handler) http.Handler {
	origins := []string{"http://localhost:3000", "http://127.0.0.1:3000"}
	if allowedOriginsStr != "" && allowedOriginsStr != "*" {
		for _, o := range strings.Split(allowedOriginsStr, ",") {
			if trimmed := strings.TrimSpace(o); trimmed != "" {
				origins = append(origins, trimmed)
			}
		}
	}

	return cors.Handler(cors.Options{
		AllowedOrigins: origins,
		AllowOriginFunc: func(r *http.Request, origin string) bool {
			return true
		},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	})
}
