package api

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sync"
	"time"

	"api-coordinador/internal/models"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type sfCall struct {
	wg  sync.WaitGroup
	val uint8
}

type Server struct {
	// Persistencia
	MongoClient  *mongo.Client
	HistorialCol *mongo.Collection
	Rdb          *redis.Client

	// Seguridad
	JWTSecret []byte

	// Modelo ML Compartido
	BosqueGlobal  []*models.TreeNode
	MatrizGlobal  [3][3]int
	RWMutex       sync.RWMutex
	ModeloVersion uint64

	// Cache de Inferencias
	CacheHits   uint64
	CacheMisses uint64
	CacheErrors uint64

	// Concurrencia Singleflight
	SfGroup map[string]*sfCall
	SfMutex sync.Mutex

	// Telemetría
	ActiveSockets uint64
	NodeTelemetry sync.Map
}

func NewServer(mongoClient *mongo.Client, rdb *redis.Client, dbName, collectionName, jwtSecret string) *Server {
	var col *mongo.Collection
	if mongoClient != nil {
		col = mongoClient.Database(dbName).Collection(collectionName)
	}

	return &Server{
		MongoClient:  mongoClient,
		HistorialCol: col,
		Rdb:          rdb,
		JWTSecret:    []byte(jwtSecret),

		ModeloVersion: 1,
		SfGroup:       make(map[string]*sfCall),
	}
}

func (s *Server) SeedAdministradores() {
	if s.MongoClient == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	col := s.MongoClient.Database("cdc_diabetes").Collection("admins_nativos")
	count, err := col.CountDocuments(ctx, bson.M{})
	if err != nil || count > 0 {
		return
	}

	hashPass := fmt.Sprintf("%x", sha256.Sum256([]byte("admin123")))
	admins := []interface{}{
		bson.M{"username": "amolihua", "password": string(hashPass)},
		bson.M{"username": "iansanchez", "password": string(hashPass)},
		bson.M{"username": "joeturpo", "password": string(hashPass)},
		bson.M{"username": "jara", "password": string(hashPass)},
	}

	_, err = col.InsertMany(ctx, admins)
	if err == nil {
		fmt.Println("[API-REST] Administradores nativos sembrados en MongoDB exitosamente.")
	}
}
