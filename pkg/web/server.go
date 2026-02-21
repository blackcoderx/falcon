package web

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"time"
)

//go:embed static
var staticFiles embed.FS

// Start binds to the requested port (0 = OS-assigned), registers all routes,
// and begins serving in a background goroutine. Returns the actual bound port
// and a shutdown function that drains the server gracefully.
func Start(zapDir string, port int) (actualPort int, shutdown func(), err error) {
	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return 0, nil, fmt.Errorf("web: failed to bind port: %w", err)
	}
	actualPort = ln.Addr().(*net.TCPAddr).Port

	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		_ = ln.Close()
		return 0, nil, fmt.Errorf("web: failed to prepare static files: %w", err)
	}

	mux := http.NewServeMux()
	registerRoutes(mux, zapDir, staticFS)

	srv := &http.Server{
		Handler:      corsMiddleware(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	go func() { _ = srv.Serve(ln) }()

	shutdown = func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
	}

	return actualPort, shutdown, nil
}

// corsMiddleware adds permissive CORS headers suitable for localhost-only use.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
