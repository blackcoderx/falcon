package tools

import "sync"

// ResponseManager manages shared state between tools
// This allows tools like assert_response and extract_value to access
// the last HTTP response from http_request tool
type ResponseManager struct {
	lastHTTPResponse *HTTPResponse
	mu               sync.RWMutex
}

// NewResponseManager creates a new response manager
func NewResponseManager() *ResponseManager {
	return &ResponseManager{}
}

// SetHTTPResponse stores the last HTTP response
func (rm *ResponseManager) SetHTTPResponse(resp *HTTPResponse) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.lastHTTPResponse = resp
}

// GetHTTPResponse retrieves the last HTTP response
func (rm *ResponseManager) GetHTTPResponse() *HTTPResponse {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return rm.lastHTTPResponse
}
