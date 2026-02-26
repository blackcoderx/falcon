package web

import (
	"io/fs"
	"net/http"
)

func registerRoutes(mux *http.ServeMux, falconDir string, staticFS fs.FS) {
	h := &handlers{falconDir: falconDir}

	// Static frontend
	mux.Handle("/", http.FileServer(http.FS(staticFS)))

	// Dashboard
	mux.HandleFunc("GET /api/dashboard", h.getDashboard)

	// Config
	mux.HandleFunc("GET /api/config", h.getConfig)
	mux.HandleFunc("PUT /api/config", h.putConfig)

	// Requests
	mux.HandleFunc("GET /api/requests", h.listRequests)
	mux.HandleFunc("POST /api/requests", h.createRequest)
	mux.HandleFunc("GET /api/requests/{name}", h.getRequest)
	mux.HandleFunc("PUT /api/requests/{name}", h.putRequest)
	mux.HandleFunc("DELETE /api/requests/{name}", h.deleteRequest)

	// Environments
	mux.HandleFunc("GET /api/environments", h.listEnvironments)
	mux.HandleFunc("POST /api/environments", h.createEnvironment)
	mux.HandleFunc("GET /api/environments/{name}", h.getEnvironment)
	mux.HandleFunc("PUT /api/environments/{name}", h.putEnvironment)
	mux.HandleFunc("DELETE /api/environments/{name}", h.deleteEnvironment)

	// Memory
	mux.HandleFunc("GET /api/memory", h.listMemory)
	mux.HandleFunc("PUT /api/memory/{key}", h.putMemoryEntry)
	mux.HandleFunc("DELETE /api/memory/{key}", h.deleteMemoryEntry)

	// Variables
	mux.HandleFunc("GET /api/variables", h.listVariables)
	mux.HandleFunc("PUT /api/variables/{name}", h.putVariable)
	mux.HandleFunc("DELETE /api/variables/{name}", h.deleteVariable)

	// Read-only
	mux.HandleFunc("GET /api/history", h.listHistory)
	mux.HandleFunc("GET /api/exports", h.listExports)
	mux.HandleFunc("GET /api/exports/{name}", h.getExport)
	mux.HandleFunc("GET /api/api-graph", h.getAPIGraph)
}
