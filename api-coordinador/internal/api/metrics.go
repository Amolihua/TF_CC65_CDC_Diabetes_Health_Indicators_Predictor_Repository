package api

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"runtime"
	"sync/atomic"
	"time"

	"api-coordinador/internal/models"
)

func (s *Server) HandleMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	// Lectura de métricas
	s.RWMutex.RLock()
	matriz := s.MatrizGlobal
	s.RWMutex.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"confusion_matrix": matriz,
		"cache_hits":       atomic.LoadUint64(&s.CacheHits),
		"cache_misses":     atomic.LoadUint64(&s.CacheMisses),
		"cache_errors":     atomic.LoadUint64(&s.CacheErrors),
		"modelo_version":   atomic.LoadUint64(&s.ModeloVersion),
	})
}

func (s *Server) HandleInternalTelemetry(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		return
	}
	var metrics models.NodoMetrics
	if err := json.NewDecoder(r.Body).Decode(&metrics); err == nil {
		s.NodeTelemetry.Store(metrics.Hostname, metrics)
	}
}

func (s *Server) HandleWSMetrics(w http.ResponseWriter, r *http.Request) {
	key := r.Header.Get("Sec-WebSocket-Key")
	h := sha1.New()
	h.Write([]byte(key + "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"))
	accept := base64.StdEncoding.EncodeToString(h.Sum(nil))

	conn, bufrw, err := w.(http.Hijacker).Hijack()
	if err != nil {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}
	bufrw.WriteString("HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Accept: " + accept + "\r\n\r\n")
	bufrw.Flush()

	go func() {
		atomic.AddUint64(&s.ActiveSockets, 1)
		defer atomic.AddUint64(&s.ActiveSockets, ^uint64(0))
		defer conn.Close()
		ticker := time.NewTicker(50 * time.Millisecond)
		defer ticker.Stop()

		for range ticker.C {
			s.RWMutex.RLock()
			matriz := s.MatrizGlobal
			s.RWMutex.RUnlock()

			var m runtime.MemStats
			runtime.ReadMemStats(&m)

			// Recolectar nodos
			var nodosList []models.NodoMetrics
			s.NodeTelemetry.Range(func(key, value interface{}) bool {
				nodosList = append(nodosList, value.(models.NodoMetrics))
				return true
			})

			payload, _ := json.Marshal(map[string]interface{}{
				"cache_hits":     atomic.LoadUint64(&s.CacheHits),
				"cache_misses":   atomic.LoadUint64(&s.CacheMisses),
				"cache_errors":   atomic.LoadUint64(&s.CacheErrors),
				"modelo_version": atomic.LoadUint64(&s.ModeloVersion),
				"nodos_activos":  atomic.LoadUint64(&s.ActiveSockets),
				"cpu_goroutines": runtime.NumGoroutine(),
				"ram_sys_mb":     m.Sys / 1024 / 1024,
				"matriz":         matriz,
				"cluster_nodes":  nodosList,
			})

			size := len(payload)
			var header []byte
			if size <= 125 {
				header = []byte{0x81, byte(size)}
			} else {
				header = []byte{0x81, 126, byte(size >> 8), byte(size & 255)}
			}

			// Intercepción de error --> Desconexión limpia del cliente
			if _, err := conn.Write(append(header, payload...)); err != nil {
				return
			}
		}
	}()
}
