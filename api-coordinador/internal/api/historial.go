package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"api-coordinador/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (s *Server) HandleHistorial(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	if s.HistorialCol == nil {
		http.Error(w, "Colección de historial no disponible", http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "created_at", Value: -1}})
	findOptions.SetLimit(50)

	cursor, err := s.HistorialCol.Find(ctx, bson.D{}, findOptions)
	if err != nil {
		http.Error(w, "Error al consultar historial", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var resultados []models.HistorialPredictivo
	if err = cursor.All(ctx, &resultados); err != nil {
		http.Error(w, "Error al decodificar historial", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resultados)
}

func (s *Server) GuardarHistorialEnMongo(p models.PerfilPaciente, diagnosis uint8, email string) {
	if s.HistorialCol == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	doc := models.HistorialPredictivo{
		Perfil:    p,
		Diagnosis: diagnosis,
		Email:     email,
		CreatedAt: time.Now(),
	}
	_, _ = s.HistorialCol.InsertOne(ctx, doc)
}
