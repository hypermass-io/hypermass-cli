package synclock

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hypermass-cli/config"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

type SyncLock struct {
	PID          int    `yaml:"pid"`
	Port         int    `yaml:"port"`
	ControlToken string `yaml:"control_token"`
	StartedAt    string `yaml:"started_at"`
}

type ControlServer struct {
	Token  string
	Port   int
	Mux    *http.ServeMux
	Bus    *CommandBus
	server *http.Server
}

func NewControlServer() (*ControlServer, error) {
	b := make([]byte, 16)
	rand.Read(b)
	token := hex.EncodeToString(b)

	return &ControlServer{
		Token: token,
		Mux:   http.NewServeMux(),
	}, nil
}

// Start Starts the control server and handle commands
func (s *ControlServer) Start() error {
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return err
	}
	s.Port = listener.Addr().(*net.TCPAddr).Port

	lock := SyncLock{
		PID:          os.Getpid(),
		Port:         s.Port,
		ControlToken: s.Token,
		StartedAt:    time.Now().Format(time.RFC3339),
	}

	data, _ := yaml.Marshal(lock)
	// 0600 = Read/Write for owner ONLY
	if err := os.WriteFile(filepath.Join(config.CreateOrGetConfigPath(), "sync-lock.yml"), data, 0600); err != nil {
		return fmt.Errorf("failed to write lockfile: %w", err)
	}

	s.Mux.HandleFunc("/ping", s.authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		//Used to test that the Control Server is running
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "pong")
	}))

	s.registerRoutes()

	s.server = &http.Server{Handler: s.Mux}

	go func() {
		fmt.Printf("Control Server running on port %d \n", s.Port)
		err := s.server.Serve(listener)
		if err != nil {
			log.Fatalf("control port registration failed, %v", err)
		}
	}()

	return nil
}

// authMiddleware filters only permitted commands with the correct credentials
func (s *ControlServer) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X-Hypermass-Token")
		if token != s.Token {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

// Register the universal command handler in ControlServer.Start()
func (s *ControlServer) registerRoutes() {
	s.Mux.HandleFunc("/cmd", s.authMiddleware(func(responseWriter http.ResponseWriter, request *http.Request) {
		// e.g. the action, e.g. "replay"
		action := request.URL.Query().Get("action")
		if action == "" {
			s.jsonError(responseWriter, "missing action parameter", http.StatusBadRequest)
			return
		}

		// Map all other query parameters into a simple string map
		params := make(map[string]string)
		for k, v := range request.URL.Query() {
			if k != "action" && len(v) > 0 {
				params[k] = v[0]
			}
		}

		// Dispatch to the Bus and get the response from the worker
		req := CommandRequest{
			Command: action,
			Params:  params,
		}

		response := s.Bus.Dispatch(req)

		// Send the worker's response back to the CLI as JSON
		responseWriter.Header().Set("Content-Type", "application/json")
		json.NewEncoder(responseWriter).Encode(response)
	}))
}

// Helper for quick internal errors
func (s *ControlServer) jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(CommandResponse{
		Success: false,
		Message: msg,
	})
}
