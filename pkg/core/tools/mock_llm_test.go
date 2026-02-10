package tools

import (
	"github.com/blackcoderx/zap/pkg/llm"
)

type MockLLMClient struct {
	Response string
	Error    error
}

func (m *MockLLMClient) Chat(messages []llm.Message) (string, error) {
	return m.Response, m.Error
}

func (m *MockLLMClient) ChatStream(messages []llm.Message, callback llm.StreamCallback) (string, error) {
	if m.Response != "" {
		callback(m.Response)
	}
	return m.Response, m.Error
}

func (m *MockLLMClient) CheckConnection() error {
	return nil
}

func (m *MockLLMClient) GetModel() string {
	return "mock-model"
}
