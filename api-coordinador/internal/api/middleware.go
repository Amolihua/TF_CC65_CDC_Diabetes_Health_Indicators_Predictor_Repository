package api

import (
	"context"
	"net/http"
	"strings"
)

// CorsMiddleware inyecta CORS a las respuestas
func (s *Server) CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// JWTMiddleware valida la sesión en Redis
func (s *Server) JWTMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, `{"error":"Acceso denegado: Token requerido"}`, http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Validar sesión en Redis usando el cliente del Server
		username, err := s.Rdb.Get(context.Background(), "session:"+tokenString).Result()
		if err != nil || username == "" {
			http.Error(w, `{"error":"Acceso denegado: Sesión inválida o expirada"}`, http.StatusUnauthorized)
			return
		}

		// Sesión válida, inyectar el usuario en el contexto
		ctx := context.WithValue(r.Context(), "username", username)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
