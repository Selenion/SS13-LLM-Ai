package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

var (
	ollamaURL   = getEnv("OLLAMA_URL", "http://localhost:11434")
	ollamaModel = getEnv("OLLAMA_MODEL", "qwen3:4b-instruct-2507-q4_K_M")
	listenAddr  = getEnv("LISTEN_ADDR", ":8000")
	client      *OllamaClient
)

func main() {
	client = NewOllamaClient(ollamaURL, ollamaModel)
	log.SetFlags(log.Ltime | log.Lmicroseconds)
	log.Printf("--- SS13 AI BACKEND STARTED [%s] ---", ollamaModel)

	// Оборачиваем в Debug + Recovery
	http.HandleFunc("/fax/analyze", wrapDebug(handleFaxAnalyze))
	http.HandleFunc("/fax/reply", wrapDebug(handleFaxReply))

	log.Printf("Listening on %s", listenAddr)
	log.Fatal(http.ListenAndServe(listenAddr, nil))
}

// Middleware
func wrapDebug(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, _ := io.ReadAll(r.Body)
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		log.Printf("INCOMING: %s %s", r.Method, r.URL.Path)

		next(w, r)
	}
}

// Умная читалка: Body -> Form -> URL Query
func readRequestData(r *http.Request, target interface{}) error {
	// 1. Сначала пробуем распарсить тело как сырой JSON (для будущего)
	// Мы читаем тело в буфер, чтобы если там не JSON, можно было прочитать снова
	bodyBytes, _ := io.ReadAll(r.Body)
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	if len(bodyBytes) > 0 {
		if err := json.Unmarshal(bodyBytes, target); err == nil {
			return nil
		}
	}

	// 2. Если не вышло, парсим как Форму (POST Form или URL Query)
	// ParseForm делает всю грязную работу за нас
	if err := r.ParseForm(); err != nil {
		return err
	}

	// Ищем параметр "data"
	dataStr := r.Form.Get("data")
	if dataStr == "" {
		return fmt.Errorf("no 'data' param found in Request")
	}

	// Декодируем JSON из строки
	return json.Unmarshal([]byte(dataStr), target)
}

// --- Handlers ---

func handleFaxAnalyze(w http.ResponseWriter, r *http.Request) {
	var req FaxRequest
	if err := readRequestData(r, &req); err != nil {
		log.Printf("Bad Request: %v", err)
		writeError(w, http.StatusBadRequest, "Invalid Data: "+err.Error())
		return
	}

	start := time.Now()
	resp, err := client.AnalyzeFax(req)
	if err != nil {
		log.Printf("LLM Error: %v", err)
		writeError(w, http.StatusServiceUnavailable, err.Error())
		return
	}

	resp.LatencyMs = time.Since(start).Milliseconds()
	writeJSON(w, http.StatusOK, resp)
}

func handleFaxReply(w http.ResponseWriter, r *http.Request) {
	var req FaxReplyRequest
	if err := readRequestData(r, &req); err != nil {
		log.Printf("Bad Request: %v", err)
		writeError(w, http.StatusBadRequest, "Invalid Data: "+err.Error())
		return
	}

	start := time.Now()
	draft, err := client.GenerateReply(req)
	if err != nil {
		log.Printf("LLM Error: %v", err)
		writeError(w, http.StatusServiceUnavailable, err.Error())
		return
	}

	resp := FaxReplyResponse{
		Draft:     draft,
		LatencyMs: time.Since(start).Milliseconds(),
	}
	writeJSON(w, http.StatusOK, resp)
}

// --- Utils ---

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
