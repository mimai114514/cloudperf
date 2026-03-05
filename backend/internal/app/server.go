package app

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"cloudperf/backend/internal/auth"
	"cloudperf/backend/internal/db"
	"cloudperf/backend/internal/orchestrator"
	"cloudperf/backend/internal/ws"
)

type Config struct {
	Addr          string
	PostgresDSN   string
	SessionTTL    time.Duration
	SecureCookie  bool
	AdminUser     string
	AdminPassword string
}

func LoadConfig() Config {
	secureCookie := true
	if strings.EqualFold(os.Getenv("COOKIE_SECURE"), "false") {
		secureCookie = false
	}
	ttl := 12 * time.Hour
	if raw := os.Getenv("SESSION_TTL_HOURS"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			ttl = time.Duration(n) * time.Hour
		}
	}
	return Config{
		Addr:          getEnv("BACKEND_ADDR", ":8080"),
		PostgresDSN:   getEnv("POSTGRES_DSN", "postgres://cloudperf:cloudperf@127.0.0.1:5432/cloudperf?sslmode=disable"),
		SessionTTL:    ttl,
		SecureCookie:  secureCookie,
		AdminUser:     getEnv("ADMIN_USERNAME", "admin"),
		AdminPassword: getEnv("ADMIN_PASSWORD", "admin123"),
	}
}

func getEnv(key, def string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	return v
}

type Server struct {
	cfg    Config
	store  *db.Store
	sess   *auth.SessionStore
	hub    *ws.Hub
	orch   *orchestrator.Orchestrator
	server *http.Server
}

func New(ctx context.Context, cfg Config) (*Server, error) {
	store, err := db.Open(cfg.PostgresDSN)
	if err != nil {
		return nil, err
	}
	if err := store.Migrate("migrations/001_init.sql"); err != nil {
		return nil, err
	}
	hash, err := auth.HashPassword(cfg.AdminPassword)
	if err != nil {
		return nil, err
	}
	if err := store.EnsureUser(db.NewID("user"), cfg.AdminUser, hash); err != nil {
		return nil, err
	}

	hub := ws.NewHub(store)
	orch := orchestrator.New(store, hub)
	go orch.Start(ctx)

	s := &Server{
		cfg:   cfg,
		store: store,
		sess:  auth.NewSessionStore(cfg.SessionTTL),
		hub:   hub,
		orch:  orch,
	}
	s.server = &http.Server{Addr: cfg.Addr, Handler: s.routes()}
	return s, nil
}

func (s *Server) Close() error {
	return s.store.Close()
}

func (s *Server) ListenAndServe() error {
	log.Printf("backend listening on %s", s.cfg.Addr)
	return s.server.ListenAndServe()
}

func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/v1/auth/login", s.handleLogin)
	mux.HandleFunc("POST /api/v1/auth/logout", s.withAuth(s.handleLogout))
	mux.HandleFunc("GET /api/v1/auth/me", s.withAuth(s.handleMe))

	mux.HandleFunc("POST /api/v1/nodes", s.withAuth(s.handleCreateNode))
	mux.HandleFunc("GET /api/v1/nodes", s.withAuth(s.handleListNodes))

	mux.HandleFunc("POST /api/v1/runs", s.withAuth(s.handleCreateRun))
	mux.HandleFunc("GET /api/v1/runs/{id}", s.withAuth(s.handleGetRun))
	mux.HandleFunc("GET /api/v1/runs/{id}/pairs", s.withAuth(s.handleGetRunPairs))

	mux.HandleFunc("GET /api/v1/results", s.withAuth(s.handleListResults))
	mux.HandleFunc("GET /api/v1/results/export.csv", s.withAuth(s.handleExportCSV))

	mux.HandleFunc("GET /agent/v1/ws", s.hub.HandleWS)

	return withCORS(mux)
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PATCH,OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) withAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("cloudperf_session")
		if err != nil || c.Value == "" {
			writeErr(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		if _, ok := s.sess.Get(c.Value); !ok {
			writeErr(w, http.StatusUnauthorized, "invalid session")
			return
		}
		next(w, r)
	}
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	u, err := s.store.GetUserByUsername(req.Username)
	if err != nil || u == nil {
		writeErr(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if !auth.CheckPassword(req.Password, u.PasswordHash) {
		writeErr(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	tok, exp := s.sess.Create(u.ID)
	http.SetCookie(w, &http.Cookie{
		Name:     "cloudperf_session",
		Value:    tok,
		Path:     "/",
		HttpOnly: true,
		Secure:   s.cfg.SecureCookie,
		SameSite: http.SameSiteLaxMode,
		Expires:  exp,
	})
	writeJSON(w, http.StatusOK, map[string]any{"id": u.ID, "username": u.Username})
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("cloudperf_session")
	if err == nil {
		s.sess.Delete(c.Value)
	}
	http.SetCookie(w, &http.Cookie{Name: "cloudperf_session", Value: "", Path: "/", MaxAge: -1, HttpOnly: true, Secure: s.cfg.SecureCookie, SameSite: http.SameSiteLaxMode})
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	c, _ := r.Cookie("cloudperf_session")
	sess, _ := s.sess.Get(c.Value)
	writeJSON(w, http.StatusOK, map[string]any{"user_id": sess.UserID})
}

func (s *Server) handleCreateNode(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Name) == "" {
		writeErr(w, http.StatusBadRequest, "name is required")
		return
	}
	nodeID := db.NewID("node")
	token := auth.NewToken()
	node, err := s.store.CreateNode(nodeID, strings.TrimSpace(req.Name), db.HashToken(token))
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"node": node, "token": token})
}

func (s *Server) handleListNodes(w http.ResponseWriter, r *http.Request) {
	nodes, err := s.store.ListNodes()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": nodes})
}

func (s *Server) handleCreateRun(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Mode     string         `json:"mode"`
		Protocol string         `json:"protocol"`
		Params   map[string]any `json:"params"`
		Pairs    []db.PairSpec  `json:"pairs"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.Mode != "one_to_one" && req.Mode != "many_to_many" {
		writeErr(w, http.StatusBadRequest, "mode must be one_to_one or many_to_many")
		return
	}
	if req.Protocol != "tcp" && req.Protocol != "udp" {
		writeErr(w, http.StatusBadRequest, "protocol must be tcp or udp")
		return
	}
	if req.Params == nil {
		req.Params = map[string]any{}
	}
	fillDefaults(req.Params)
	if err := validateParams(req.Protocol, req.Params); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := orchestrator.ValidatePairs(req.Pairs); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	for _, p := range req.Pairs {
		if p.SourceNodeID == p.TargetNodeID {
			writeErr(w, http.StatusBadRequest, "source and target must be different")
			return
		}
		n1, err := s.store.GetNodeByID(p.SourceNodeID)
		if err != nil || n1 == nil {
			writeErr(w, http.StatusBadRequest, "unknown source node: "+p.SourceNodeID)
			return
		}
		n2, err := s.store.GetNodeByID(p.TargetNodeID)
		if err != nil || n2 == nil {
			writeErr(w, http.StatusBadRequest, "unknown target node: "+p.TargetNodeID)
			return
		}
	}

	runID := db.NewID("run")
	if err := s.store.CreateRun(runID, req.Mode, req.Protocol, req.Params, req.Pairs); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.orch.Enqueue(runID)
	writeJSON(w, http.StatusCreated, map[string]any{"run_id": runID})
}

func fillDefaults(params map[string]any) {
	if _, ok := params["duration"]; !ok {
		params["duration"] = 10
	}
	if _, ok := params["parallel_streams"]; !ok {
		params["parallel_streams"] = 1
	}
	if _, ok := params["udp_bandwidth"]; !ok {
		params["udp_bandwidth"] = "100M"
	}
}

func validateParams(protocol string, params map[string]any) error {
	duration := toInt(params["duration"], 10)
	if duration < 1 || duration > 300 {
		return errors.New("duration must be within 1..300")
	}
	parallel := toInt(params["parallel_streams"], 1)
	if parallel < 1 || parallel > 20 {
		return errors.New("parallel_streams must be within 1..20")
	}
	if protocol == "udp" {
		bw, _ := params["udp_bandwidth"].(string)
		if strings.TrimSpace(bw) == "" {
			return errors.New("udp_bandwidth is required for udp")
		}
	}
	return nil
}

func toInt(v any, def int) int {
	switch n := v.(type) {
	case int:
		return n
	case float64:
		return int(n)
	case string:
		i, err := strconv.Atoi(n)
		if err != nil {
			return def
		}
		return i
	default:
		return def
	}
}

func (s *Server) handleGetRun(w http.ResponseWriter, r *http.Request) {
	run, err := s.store.GetRun(r.PathValue("id"))
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if run == nil {
		writeErr(w, http.StatusNotFound, "run not found")
		return
	}
	writeJSON(w, http.StatusOK, run)
}

func (s *Server) handleGetRunPairs(w http.ResponseWriter, r *http.Request) {
	pairs, err := s.store.ListRunPairs(r.PathValue("id"))
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": pairs})
}

func (s *Server) handleListResults(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit := toInt(q.Get("limit"), 200)
	items, err := s.store.ListResults(q.Get("run_id"), q.Get("from"), q.Get("to"), limit)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (s *Server) handleExportCSV(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	items, err := s.store.ListResults(q.Get("run_id"), q.Get("from"), q.Get("to"), 1000)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	filename := fmt.Sprintf("cloudperf-results-%s.csv", time.Now().Format("20060102-150405"))
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename="+filename)

	writer := csv.NewWriter(w)
	_ = writer.Write([]string{"result_id", "pair_id", "protocol", "created_at", "metrics_json"})
	for _, it := range items {
		m, _ := json.Marshal(it.Metrics)
		_ = writer.Write([]string{it.ID, it.PairID, it.Protocol, it.CreatedAt.Format(time.RFC3339), string(m)})
	}
	writer.Flush()
}

func writeErr(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]any{"error": msg})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("write json error: %v", err)
	}
}
