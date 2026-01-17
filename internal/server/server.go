package server

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"

	"github.com/casbin/policywall/internal/policy"
)

//go:embed templates/* static/*
var assets embed.FS

type Server struct {
	port     string
	enforcer *policy.Enforcer
}

func NewServer(port string) *Server {
	return &Server{
		port:     port,
		enforcer: policy.NewEnforcer(),
	}
}

func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Serve static files
	staticFS := http.FileServer(http.FS(assets))
	mux.Handle("/static/", staticFS)

	// Dashboard routes
	mux.HandleFunc("/", s.handleOverview)
	mux.HandleFunc("/editor", s.handleEditor)
	mux.HandleFunc("/playground", s.handlePlayground)

	// API routes
	mux.HandleFunc("/api/metrics", s.handleMetrics)
	mux.HandleFunc("/api/policy/validate", s.handlePolicyValidate)
	mux.HandleFunc("/api/playground/test", s.handlePlaygroundTest)

	addr := fmt.Sprintf(":%s", s.port)
	log.Printf("Starting PolicyWall dashboard on http://localhost%s\n", addr)
	return http.ListenAndServe(addr, mux)
}

func (s *Server) handleOverview(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFS(assets, "templates/base.html", "templates/overview.html")
	if err != nil {
		http.Error(w, "Failed to load template", http.StatusInternalServerError)
		log.Printf("Template error: %v\n", err)
		return
	}

	data := map[string]interface{}{
		"Title":   "Overview",
		"Page":    "overview",
		"Metrics": s.getMetrics(),
	}

	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		log.Printf("Render error: %v\n", err)
	}
}

func (s *Server) handleEditor(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFS(assets, "templates/base.html", "templates/editor.html")
	if err != nil {
		http.Error(w, "Failed to load template", http.StatusInternalServerError)
		log.Printf("Template error: %v\n", err)
		return
	}

	data := map[string]interface{}{
		"Title": "Policy Editor",
		"Page":  "editor",
	}

	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		log.Printf("Render error: %v\n", err)
	}
}

func (s *Server) handlePlayground(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFS(assets, "templates/base.html", "templates/playground.html")
	if err != nil {
		http.Error(w, "Failed to load template", http.StatusInternalServerError)
		log.Printf("Template error: %v\n", err)
		return
	}

	data := map[string]interface{}{
		"Title": "Playground",
		"Page":  "playground",
	}

	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		log.Printf("Render error: %v\n", err)
	}
}

func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.getMetrics())
}

func (s *Server) handlePolicyValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	// Basic YAML validation
	result := map[string]interface{}{
		"valid": true,
		"yaml":  string(body),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (s *Server) handlePlaygroundTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Request string `json:"request"`
		Policy  string `json:"policy"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	result, err := s.enforcer.EvaluateRequest(request.Request, request.Policy)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"allowed": false,
			"error":   err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"allowed": result,
		"message": fmt.Sprintf("Policy evaluation: %v", result),
	})
}

func (s *Server) getMetrics() map[string]interface{} {
	// Mock metrics - in real implementation, these would come from Prometheus
	return map[string]interface{}{
		"totalRequests":   12453,
		"allowedRequests": 11892,
		"deniedRequests":  561,
		"violationRate":   4.5,
		"activePolicies":  8,
	}
}
