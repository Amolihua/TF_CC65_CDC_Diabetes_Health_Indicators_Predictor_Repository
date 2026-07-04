package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"api-coordinador/internal/api"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "secreto-super-seguro-pc4"
	}

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var mongoClient *mongo.Client
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		fmt.Printf("[CRÍTICO] Fallo al conectar con MongoDB: %v\n", err)
	} else {
		mongoClient = client
		fmt.Println("[API-REST] Conexión establecida con MongoDB en", mongoURI)
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		fmt.Printf("[API-REST] ADVERTENCIA: Redis inaccesible en %s: %v\n", redisAddr, err)
	} else {
		fmt.Printf("[API-REST] Ping exitoso a Redis en %s\n", redisAddr)
	}

	// Instanciar el Servidor (Inyección de Dependencias)
	server := api.NewServer(mongoClient, rdb, "cdc_diabetes", "predicciones", secret)

	// Sembrar base de datos
	server.SeedAdministradores()

	// Enrutador nativo
	mux := http.NewServeMux()
	mux.HandleFunc("/api/login", server.HandleLogin)
	mux.HandleFunc("/api/train", server.JWTMiddleware(server.HandleTrain))
	mux.HandleFunc("/api/predict", server.HandlePredict)
	mux.HandleFunc("/api/metrics", server.JWTMiddleware(server.HandleMetrics))
	mux.HandleFunc("/api/historial", server.JWTMiddleware(server.HandleHistorial))
	mux.HandleFunc("/api/ws/metrics", server.HandleWSMetrics)
	mux.HandleFunc("/api/internal/telemetry", server.HandleInternalTelemetry)

	fmt.Println("[API-REST] Servidor HTTP de escucha perpetua iniciado en :8080")
	if err := http.ListenAndServe(":8080", server.CorsMiddleware(mux)); err != nil {
		fmt.Printf("[CRÍTICO] Fallo en el servidor HTTP: %v\n", err)
	}
}
