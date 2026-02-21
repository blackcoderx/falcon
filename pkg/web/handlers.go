package web

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/blackcoderx/zap/pkg/core"
	"github.com/blackcoderx/zap/pkg/storage"
)

type handlers struct {
	zapDir string
}

// writeJSON serializes v as JSON with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError sends a {"error": msg} JSON response.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// readJSON decodes the request body into v.
func readJSON(r *http.Request, v any) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}

// safeName sanitizes a URL path parameter to prevent path traversal.
func safeName(raw string) (string, bool) {
	name := filepath.Base(raw)
	if strings.ContainsAny(name, `/\`) || name == "." || name == ".." {
		return "", false
	}
	return name, true
}

// ── Dashboard ──────────────────────────────────────────────────────────────

func (h *handlers) getDashboard(w http.ResponseWriter, r *http.Request) {
	manifest, err := readManifest(h.zapDir)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	cfg, err := readConfig(h.zapDir)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"manifest": manifest,
		"config_summary": map[string]string{
			"provider":  cfg.Provider,
			"model":     cfg.DefaultModel,
			"framework": cfg.Framework,
		},
	})
}

// ── Config ─────────────────────────────────────────────────────────────────

func (h *handlers) getConfig(w http.ResponseWriter, r *http.Request) {
	cfg, err := readConfig(h.zapDir)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, cfg)
}

func (h *handlers) putConfig(w http.ResponseWriter, r *http.Request) {
	var cfg core.Config
	if err := readJSON(r, &cfg); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if err := writeConfig(h.zapDir, &cfg); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ── Requests ───────────────────────────────────────────────────────────────

func (h *handlers) listRequests(w http.ResponseWriter, r *http.Request) {
	names, err := listRequestNames(h.zapDir)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, names)
}

func (h *handlers) getRequest(w http.ResponseWriter, r *http.Request) {
	name, ok := safeName(r.PathValue("name"))
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid name")
		return
	}
	req, err := readRequest(h.zapDir, name)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, req)
}

func (h *handlers) putRequest(w http.ResponseWriter, r *http.Request) {
	name, ok := safeName(r.PathValue("name"))
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid name")
		return
	}
	var req storage.Request
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if err := writeRequest(h.zapDir, name, &req); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *handlers) createRequest(w http.ResponseWriter, r *http.Request) {
	var req storage.Request
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	name, ok := safeName(req.Name)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid name")
		return
	}
	if err := writeRequest(h.zapDir, name, &req); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"status": "created", "name": name})
}

func (h *handlers) deleteRequest(w http.ResponseWriter, r *http.Request) {
	name, ok := safeName(r.PathValue("name"))
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid name")
		return
	}
	if err := deleteRequestFile(h.zapDir, name); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ── Environments ───────────────────────────────────────────────────────────

func (h *handlers) listEnvironments(w http.ResponseWriter, r *http.Request) {
	names, err := listEnvironmentNames(h.zapDir)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, names)
}

func (h *handlers) getEnvironment(w http.ResponseWriter, r *http.Request) {
	name, ok := safeName(r.PathValue("name"))
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid name")
		return
	}
	env, err := readEnvironment(h.zapDir, name)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, env)
}

func (h *handlers) putEnvironment(w http.ResponseWriter, r *http.Request) {
	name, ok := safeName(r.PathValue("name"))
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid name")
		return
	}
	var vars map[string]string
	if err := readJSON(r, &vars); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if err := writeEnvironment(h.zapDir, name, vars); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *handlers) createEnvironment(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name string            `json:"name"`
		Vars map[string]string `json:"vars"`
	}
	if err := readJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if body.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	name, ok := safeName(body.Name)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid name")
		return
	}
	if body.Vars == nil {
		body.Vars = map[string]string{}
	}
	if err := writeEnvironment(h.zapDir, name, body.Vars); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"status": "created", "name": name})
}

func (h *handlers) deleteEnvironment(w http.ResponseWriter, r *http.Request) {
	name, ok := safeName(r.PathValue("name"))
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid name")
		return
	}
	if err := deleteEnvironmentFile(h.zapDir, name); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ── Memory ─────────────────────────────────────────────────────────────────

func (h *handlers) listMemory(w http.ResponseWriter, r *http.Request) {
	entries, err := readMemory(h.zapDir)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, entries)
}

func (h *handlers) putMemoryEntry(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if key == "" {
		writeError(w, http.StatusBadRequest, "key is required")
		return
	}
	var body struct {
		Value    string `json:"value"`
		Category string `json:"category"`
	}
	if err := readJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}

	entries, err := readMemory(h.zapDir)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	now := time.Now().UTC().Format(time.RFC3339)
	found := false
	for i, e := range entries {
		if e.Key == key {
			entries[i].Value = body.Value
			entries[i].Category = body.Category
			entries[i].Timestamp = now
			found = true
			break
		}
	}
	if !found {
		entries = append(entries, core.MemoryEntry{
			Key:       key,
			Value:     body.Value,
			Category:  body.Category,
			Timestamp: now,
			Source:    "web-ui",
		})
	}

	if err := writeMemory(h.zapDir, entries); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *handlers) deleteMemoryEntry(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if key == "" {
		writeError(w, http.StatusBadRequest, "key is required")
		return
	}

	entries, err := readMemory(h.zapDir)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	filtered := entries[:0]
	for _, e := range entries {
		if e.Key != key {
			filtered = append(filtered, e)
		}
	}

	if err := writeMemory(h.zapDir, filtered); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ── Variables ──────────────────────────────────────────────────────────────

func (h *handlers) listVariables(w http.ResponseWriter, r *http.Request) {
	vars, err := readVariables(h.zapDir)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, vars)
}

func (h *handlers) putVariable(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	var body struct {
		Value string `json:"value"`
	}
	if err := readJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}

	vars, err := readVariables(h.zapDir)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	vars[name] = body.Value
	if err := writeVariables(h.zapDir, vars); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *handlers) deleteVariable(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	vars, err := readVariables(h.zapDir)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	delete(vars, name)
	if err := writeVariables(h.zapDir, vars); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ── Read-only ──────────────────────────────────────────────────────────────

func (h *handlers) listHistory(w http.ResponseWriter, r *http.Request) {
	sessions, err := readHistory(h.zapDir)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, sessions)
}

func (h *handlers) listExports(w http.ResponseWriter, r *http.Request) {
	names, err := listExportFiles(h.zapDir)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, names)
}

func (h *handlers) getExport(w http.ResponseWriter, r *http.Request) {
	name, ok := safeName(r.PathValue("name"))
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid name")
		return
	}
	data, err := readExportFile(h.zapDir, name)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if strings.HasSuffix(name, ".json") {
		w.Header().Set("Content-Type", "application/json")
	} else {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func (h *handlers) getAPIGraph(w http.ResponseWriter, r *http.Request) {
	data, err := readAPIGraph(h.zapDir)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}
