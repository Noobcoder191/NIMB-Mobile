package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Config holds the app configuration
type Config struct {
	ShowReasoning    bool    `json:"showReasoning"`
	EnableThinking   bool    `json:"enableThinking"`
	LogRequests      bool    `json:"logRequests"`
	ContextSize      int     `json:"contextSize"`
	MaxTokens        int     `json:"maxTokens"`
	Temperature      float64 `json:"temperature"`
	StreamingEnabled bool    `json:"streamingEnabled"`
	CurrentModel     string  `json:"currentModel"`
	APIKey           string  `json:"apiKey,omitempty"`
}

// Stats holds usage statistics
type Stats struct {
	MessageCount     int         `json:"messageCount"`
	PromptTokens     int         `json:"promptTokens"`
	CompletionTokens int         `json:"completionTokens"`
	TotalTokens      int         `json:"totalTokens"`
	ErrorCount       int         `json:"errorCount"`
	LastRequestTime  string      `json:"lastRequestTime"`
	StartTime        string      `json:"startTime"`
	ErrorLog         []ErrorItem `json:"errorLog"`
}

// ErrorItem represents an error log entry
type ErrorItem struct {
	Timestamp string `json:"timestamp"`
	Message   string `json:"message"`
	Code      int    `json:"code"`
}

// TunnelState holds cloudflare tunnel state
type TunnelState struct {
	URL     string `json:"url"`
	Status  string `json:"status"`
	process *exec.Cmd
	mu      sync.Mutex
}

// App struct
type App struct {
	config      Config
	stats       Stats
	tunnel      TunnelState
	startTime   time.Time
	settingsDir string
	mu          sync.RWMutex
}

// NewApp creates a new App
func NewApp() *App {
	homeDir, _ := os.UserHomeDir()
	settingsDir := filepath.Join(homeDir, ".nimb")
	os.MkdirAll(settingsDir, 0755)

	app := &App{
		startTime:   time.Now(),
		settingsDir: settingsDir,
		config: Config{
			ShowReasoning:    false,
			EnableThinking:   false,
			LogRequests:      true,
			ContextSize:      128000,
			MaxTokens:        0,
			Temperature:      0.7,
			StreamingEnabled: true,
			CurrentModel:     "deepseek-ai/deepseek-v3.2",
		},
		stats: Stats{
			StartTime: time.Now().Format(time.RFC3339),
			ErrorLog:  []ErrorItem{},
		},
		tunnel: TunnelState{
			Status: "stopped",
		},
	}

	app.loadSettings()
	return app
}

// Settings persistence
func (a *App) loadSettings() {
	path := filepath.Join(a.settingsDir, "settings.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	var saved Config
	if err := json.Unmarshal(data, &saved); err != nil {
		return
	}

	a.mu.Lock()
	a.config = saved
	a.mu.Unlock()
	log.Println("Loaded settings from:", path)
}

func (a *App) saveSettings() error {
	a.mu.RLock()
	data, err := json.MarshalIndent(a.config, "", "  ")
	a.mu.RUnlock()
	if err != nil {
		return err
	}

	path := filepath.Join(a.settingsDir, "settings.json")
	return os.WriteFile(path, data, 0644)
}

// GetHealth returns current health status
func (a *App) GetHealth() map[string]interface{} {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return map[string]interface{}{
		"status":             "ok",
		"service":            "NIMB Mobile",
		"model":              a.config.CurrentModel,
		"api_key_configured": a.config.APIKey != "",
		"config":             a.config,
		"stats":              a.stats,
		"tunnel": map[string]string{
			"url":    a.tunnel.URL,
			"status": a.tunnel.Status,
		},
		"uptime":        int(time.Since(a.startTime).Seconds()),
		"setupComplete": a.config.APIKey != "",
	}
}

// StartTunnel starts cloudflare tunnel
func (a *App) StartTunnel() map[string]interface{} {
	a.tunnel.mu.Lock()
	defer a.tunnel.mu.Unlock()

	if a.tunnel.Status == "running" {
		return map[string]interface{}{
			"success": true,
			"url":     a.tunnel.URL,
			"status":  "running",
		}
	}

	// Find cloudflared binary
	var cfPath string
	if runtime.GOOS == "windows" {
		exePath, _ := os.Executable()
		exeDir := filepath.Dir(exePath)
		cfPath = filepath.Join(exeDir, "cloudflared.exe")
		if _, err := os.Stat(cfPath); os.IsNotExist(err) {
			return map[string]interface{}{
				"success": false,
				"error":   "cloudflared not found. Place it next to the executable.",
			}
		}
	} else {
		// On Linux/Termux, use absolute path to avoid exec.LookPath syscall crash
		// exec.Command internally calls LookPath which uses faccessat2 - not available on Android
		termuxPath := "/data/data/com.termux/files/usr/bin/cloudflared"
		if _, err := os.Stat(termuxPath); err == nil {
			cfPath = termuxPath
		} else {
			// Fallback to common Linux paths
			for _, p := range []string{"/usr/bin/cloudflared", "/usr/local/bin/cloudflared"} {
				if _, err := os.Stat(p); err == nil {
					cfPath = p
					break
				}
			}
			if cfPath == "" {
				return map[string]interface{}{
					"success": false,
					"error":   "cloudflared not found. Install with: pkg install cloudflared",
				}
			}
		}
		log.Println("Using cloudflared at:", cfPath)
	}

	a.tunnel.Status = "starting"

	cmd := exec.Command(cfPath, "tunnel", "--url", "http://localhost:3000")

	// Capture both stdout and stderr
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		a.tunnel.Status = "stopped"
		return map[string]interface{}{
			"success": false,
			"error":   "Failed to start cloudflared: " + err.Error(),
		}
	}

	a.tunnel.process = cmd

	// Helper to scan output for tunnel URL
	scanForURL := func(output string) {
		if strings.Contains(output, "trycloudflare.com") {
			start := strings.Index(output, "https://")
			if start != -1 {
				end := strings.Index(output[start:], " ")
				if end == -1 {
					end = len(output) - start
				}
				url := strings.TrimSpace(output[start : start+end])
				if strings.HasSuffix(url, ".") {
					url = url[:len(url)-1]
				}
				a.tunnel.mu.Lock()
				a.tunnel.URL = url
				a.tunnel.Status = "running"
				a.tunnel.mu.Unlock()
				log.Println("Tunnel URL:", url)
			}
		}
	}

	// Read from stderr
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := stderr.Read(buf)
			if err != nil {
				break
			}
			output := string(buf[:n])
			log.Println("Cloudflared:", output)
			scanForURL(output)
		}
	}()

	// Read from stdout (cloudflared may output to either)
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := stdout.Read(buf)
			if err != nil {
				break
			}
			output := string(buf[:n])
			log.Println("Cloudflared:", output)
			scanForURL(output)
		}
	}()

	// Wait for process to exit
	go func() {
		cmd.Wait()
		a.tunnel.mu.Lock()
		a.tunnel.Status = "stopped"
		a.tunnel.URL = ""
		a.tunnel.process = nil
		a.tunnel.mu.Unlock()
	}()

	return map[string]interface{}{
		"success": true,
		"status":  "starting",
	}
}

// StopTunnel stops cloudflare tunnel
func (a *App) StopTunnel() bool {
	a.tunnel.mu.Lock()
	defer a.tunnel.mu.Unlock()

	if a.tunnel.process != nil {
		a.tunnel.process.Process.Kill()
		a.tunnel.process = nil
	}
	a.tunnel.Status = "stopped"
	a.tunnel.URL = ""
	return true
}

// HTTP API Handlers

func (a *App) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(a.GetHealth())
}

func (a *App) handleHealthJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(a.GetHealth())
}

func (a *App) handleConfig(w http.ResponseWriter, r *http.Request) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(a.config)
}

func (a *App) handleSaveConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var cfg Config
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	a.mu.Lock()
	if cfg.APIKey == "" {
		cfg.APIKey = a.config.APIKey
	}
	a.config = cfg
	a.mu.Unlock()

	if err := a.saveSettings(); err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"success": false})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func (a *App) handleSetModel(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Model string `json:"model"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	a.mu.Lock()
	a.config.CurrentModel = req.Model
	a.mu.Unlock()

	success := a.saveSettings() == nil
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": success})
}

func (a *App) handleSetAPIKey(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Key string `json:"key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	a.mu.Lock()
	a.config.APIKey = req.Key
	a.mu.Unlock()

	success := a.saveSettings() == nil
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": success})
}

func (a *App) handleStats(w http.ResponseWriter, r *http.Request) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(a.stats)
}

func (a *App) handleResetStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	a.mu.Lock()
	a.stats = Stats{
		StartTime: time.Now().Format(time.RFC3339),
		ErrorLog:  []ErrorItem{},
	}
	a.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func (a *App) handleStartTunnel(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	result := a.StartTunnel()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (a *App) handleStopTunnel(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	a.StopTunnel()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func (a *App) handleTunnelStatus(w http.ResponseWriter, r *http.Request) {
	a.tunnel.mu.Lock()
	defer a.tunnel.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"url":    a.tunnel.URL,
		"status": a.tunnel.Status,
	})
}

func (a *App) handleModels(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"object":"list","data":[]}`))
}

func (a *App) handleChatCompletions(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	a.mu.RLock()
	apiKey := a.config.APIKey
	config := a.config
	a.mu.RUnlock()

	if apiKey == "" {
		a.logError("API key not configured", 500)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write([]byte(`{"error":{"message":"API key not configured","type":"configuration_error","code":500}}`))
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		a.logError(err.Error(), 400)
		http.Error(w, err.Error(), 400)
		return
	}

	var reqBody map[string]interface{}
	if err := json.Unmarshal(body, &reqBody); err != nil {
		a.logError(err.Error(), 400)
		http.Error(w, err.Error(), 400)
		return
	}

	nimReq := map[string]interface{}{
		"model":    config.CurrentModel,
		"messages": reqBody["messages"],
	}

	if temp, ok := reqBody["temperature"].(float64); ok {
		nimReq["temperature"] = temp
	} else {
		nimReq["temperature"] = config.Temperature
	}

	if maxTok, ok := reqBody["max_tokens"].(float64); ok {
		nimReq["max_tokens"] = int(maxTok)
	} else {
		nimReq["max_tokens"] = config.MaxTokens
	}

	if stream, ok := reqBody["stream"].(bool); ok {
		nimReq["stream"] = stream
	} else {
		nimReq["stream"] = config.StreamingEnabled
	}

	passthroughParams := []string{"top_p", "top_k", "frequency_penalty", "presence_penalty", "repetition_penalty", "min_p", "seed", "stop", "n", "context_length", "context_window", "truncate"}
	for _, p := range passthroughParams {
		if v, ok := reqBody[p]; ok {
			nimReq[p] = v
		}
	}

	if config.LogRequests {
		log.Printf("[NIMB] %v -> %s", reqBody["model"], config.CurrentModel)
	}

	nimBody, _ := json.Marshal(nimReq)

	// Create custom dialer with explicit DNS resolver (fixes Android IPv6 DNS issue)
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		Resolver: &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				// Force IPv4 Google DNS
				d := net.Dialer{Timeout: 10 * time.Second}
				return d.DialContext(ctx, "udp", "8.8.8.8:53")
			},
		},
	}

	transport := &http.Transport{
		DialContext:           dialer.DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: runtime.GOOS != "windows", // Skip on Android/Linux where system CAs aren't available to Go
		},
	}

	client := &http.Client{
		Timeout:   120 * time.Second,
		Transport: transport,
	}

	nimReqHTTP, _ := http.NewRequest("POST", "https://integrate.api.nvidia.com/v1/chat/completions", bytes.NewReader(nimBody))
	nimReqHTTP.Header.Set("Authorization", "Bearer "+apiKey)
	nimReqHTTP.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(nimReqHTTP)
	if err != nil {
		a.logError(err.Error(), 500)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"message": err.Error(),
				"type":    "api_error",
				"code":    500,
			},
		})
		return
	}
	defer resp.Body.Close()

	a.mu.Lock()
	a.stats.MessageCount++
	a.stats.LastRequestTime = time.Now().Format(time.RFC3339)
	a.mu.Unlock()

	isStream := nimReq["stream"].(bool)

	if isStream {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming not supported", 500)
			return
		}

		buf := make([]byte, 4096)
		for {
			n, err := resp.Body.Read(buf)
			if n > 0 {
				w.Write(buf[:n])
				flusher.Flush()
			}
			if err != nil {
				break
			}
		}
	} else {
		respBody, _ := io.ReadAll(resp.Body)

		var nimResp map[string]interface{}
		json.Unmarshal(respBody, &nimResp)

		if usage, ok := nimResp["usage"].(map[string]interface{}); ok {
			a.mu.Lock()
			if pt, ok := usage["prompt_tokens"].(float64); ok {
				a.stats.PromptTokens += int(pt)
			}
			if ct, ok := usage["completion_tokens"].(float64); ok {
				a.stats.CompletionTokens += int(ct)
			}
			if tt, ok := usage["total_tokens"].(float64); ok {
				a.stats.TotalTokens += int(tt)
			}
			a.mu.Unlock()
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		w.Write(respBody)
	}

	if config.LogRequests {
		log.Println("[NIMB] Done")
	}
}

func (a *App) logError(msg string, code int) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.stats.ErrorCount++
	a.stats.ErrorLog = append([]ErrorItem{{
		Timestamp: time.Now().Format(time.RFC3339),
		Message:   msg,
		Code:      code,
	}}, a.stats.ErrorLog...)

	if len(a.stats.ErrorLog) > 50 {
		a.stats.ErrorLog = a.stats.ErrorLog[:50]
	}
}
