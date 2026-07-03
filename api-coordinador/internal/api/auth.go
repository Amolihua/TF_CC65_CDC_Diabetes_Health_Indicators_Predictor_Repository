package api

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"api-coordinador/internal/models"

	"go.mongodb.org/mongo-driver/bson"
)

// HandleLogin procesa el endpoint público para expedir token con 24h de expiración
func (s *Server) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	var creds models.Credenciales
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Petición inválida", http.StatusBadRequest)
		return
	}

	var adminData struct {
		Username string `bson:"username"`
		Password string `bson:"password"`
	}

	if s.MongoClient == nil {
		http.Error(w, `{"error":"Servicio no disponible"}`, http.StatusInternalServerError)
		return
	}

	col := s.MongoClient.Database("cdc_diabetes").Collection("admins_nativos")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := col.FindOne(ctx, bson.M{"username": creds.Username}).Decode(&adminData)
	if err != nil {
		http.Error(w, `{"error":"Credenciales incorrectas"}`, http.StatusUnauthorized)
		return
	}

	hashInput := fmt.Sprintf("%x", sha256.Sum256([]byte(creds.Password)))
	if adminData.Password != hashInput {
		http.Error(w, `{"error":"Credenciales incorrectas"}`, http.StatusUnauthorized)
		return
	}

	// Generar Token
	tokenBytes := make([]byte, 32)
	rand.Read(tokenBytes)
	token := base64.URLEncoding.EncodeToString(tokenBytes)

	s.Rdb.Set(context.Background(), "session:"+token, creds.Username, 24*time.Hour)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}
