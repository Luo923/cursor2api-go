// Copyright (c) 2025-2026 libaxuan
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package models

import (
	"testing"
)

func TestGetStringContent(t *testing.T) {
	tests := []struct {
		name     string
		content  interface{}
		expected string
	}{
		{
			name:     "string content",
			content:  "Hello world",
			expected: "Hello world",
		},
		{
			name: "array content",
			content: []ContentPart{
				{Type: "text", Text: "Hello"},
				{Type: "text", Text: " world"},
			},
			expected: "Hello world",
		},
		{
			name:     "empty array",
			content:  []ContentPart{},
			expected: "",
		},
		{
			name:     "nil content",
			content:  nil,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := &Message{Content: tt.content}
			result := msg.GetStringContent()
			if result != tt.expected {
				t.Errorf("GetStringContent() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestToCursorMessages(t *testing.T) {
	tests := []struct {
		name              string
		messages          []Message
		systemPrompt      string
		expectedLength    int
		expectedFirstRole string
		expectedFirstMsg  string
	}{
		{
			name: "no system prompt",
			messages: []Message{
				{Role: "user", Content: "Hello"},
			},
			systemPrompt:      "",
			expectedLength:    1,
			expectedFirstRole: "user",
			expectedFirstMsg:  "Hello",
		},
		{
			name: "with system prompt, no system message",
			messages: []Message{
				{Role: "user", Content: "Hello"},
			},
			systemPrompt:      "You are a helpful assistant",
			expectedLength:    1,
			expectedFirstRole: "user",
			expectedFirstMsg:  "You are a helpful assistant\n\nHello",
		},
		{
			name: "with system prompt, has system message",
			messages: []Message{
				{Role: "system", Content: "Be helpful"},
				{Role: "user", Content: "Hello"},
			},
			systemPrompt:      "You are an AI",
			expectedLength:    1,
			expectedFirstRole: "user",
			expectedFirstMsg:  "Be helpful\n\nYou are an AI\n\nHello",
		},
		{
			name: "leading assistant message is merged into first user message",
			messages: []Message{
				{Role: "assistant", Content: "Hi there"},
				{Role: "user", Content: "Hello"},
			},
			systemPrompt:      "",
			expectedLength:    1,
			expectedFirstRole: "user",
			expectedFirstMsg:  "Hi there\n\nHello",
		},
		{
			name: "consecutive user messages are merged",
			messages: []Message{
				{Role: "user", Content: "First"},
				{Role: "user", Content: "Second"},
			},
			systemPrompt:      "",
			expectedLength:    1,
			expectedFirstRole: "user",
			expectedFirstMsg:  "First\n\nSecond",
		},
		{
			name: "system then leading assistant then user - all normalized to user",
			messages: []Message{
				{Role: "system", Content: "Be helpful"},
				{Role: "assistant", Content: "Hi"},
				{Role: "user", Content: "Hello"},
			},
			systemPrompt:      "",
			expectedLength:    1,
			expectedFirstRole: "user",
			expectedFirstMsg:  "Be helpful\n\nHi\n\nHello",
		},
		{
			name: "only non-user messages become a single user message",
			messages: []Message{
				{Role: "assistant", Content: "Hi"},
				{Role: "assistant", Content: "There"},
			},
			systemPrompt:      "",
			expectedLength:    1,
			expectedFirstRole: "user",
			expectedFirstMsg:  "Hi\n\nThere",
		},
		{
			name: "assistant replies later are also normalized to user",
			messages: []Message{
				{Role: "user", Content: "Question"},
				{Role: "assistant", Content: "Answer"},
				{Role: "user", Content: "Follow-up"},
			},
			systemPrompt:      "",
			expectedLength:    1,
			expectedFirstRole: "user",
			expectedFirstMsg:  "Question\n\nAnswer\n\nFollow-up",
		},
		{
			name: "alternating user and assistant turns collapse into one user message",
			messages: []Message{
				{Role: "user", Content: "U1"},
				{Role: "assistant", Content: "A1"},
				{Role: "user", Content: "U2"},
				{Role: "assistant", Content: "A2"},
				{Role: "user", Content: "U3"},
			},
			systemPrompt:      "",
			expectedLength:    1,
			expectedFirstRole: "user",
			expectedFirstMsg:  "U1\n\nA1\n\nU2\n\nA2\n\nU3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToCursorMessages(tt.messages, tt.systemPrompt)
			if len(result) != tt.expectedLength {
				t.Errorf("ToCursorMessages() length = %v, want %v", len(result), tt.expectedLength)
			}
			if len(result) > 0 && result[0].Role != tt.expectedFirstRole {
				t.Errorf("ToCursorMessages() first role = %v, want %v", result[0].Role, tt.expectedFirstRole)
			}
			if len(result) > 0 && result[0].Parts[0].Text != tt.expectedFirstMsg {
				t.Errorf("ToCursorMessages() first message = %v, want %v", result[0].Parts[0].Text, tt.expectedFirstMsg)
			}
		})
	}
}

func TestToCursorMessagesAllOutboundRolesBecomeUser(t *testing.T) {
	messages := []Message{
		{Role: "system", Content: "Be helpful"},
		{Role: "assistant", Content: "Hi"},
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "How can I help?"},
	}
	result := ToCursorMessages(messages, "")

	if len(result) != 1 {
		t.Fatalf("expected 1 message, got %d", len(result))
	}
	if result[0].Role != "user" {
		t.Fatalf("expected first message role=user, got %q", result[0].Role)
	}
	const want = "Be helpful\n\nHi\n\nHello\n\nHow can I help?"
	if result[0].Parts[0].Text != want {
		t.Errorf("expected merged message text %q, got %q", want, result[0].Parts[0].Text)
	}
}

func TestNewChatCompletionResponse(t *testing.T) {
	response := NewChatCompletionResponse("test-id", "gpt-4o", "Hello world", Usage{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15})

	if response.ID != "test-id" {
		t.Errorf("ID = %v, want test-id", response.ID)
	}
	if response.Model != "gpt-4o" {
		t.Errorf("Model = %v, want gpt-4o", response.Model)
	}
	if response.Choices[0].Message.Content != "Hello world" {
		t.Errorf("Content = %v, want Hello world", response.Choices[0].Message.Content)
	}
	if response.Usage.PromptTokens != 10 {
		t.Errorf("PromptTokens = %v, want 10", response.Usage.PromptTokens)
	}
}

func TestNewChatCompletionStreamResponse(t *testing.T) {
	response := NewChatCompletionStreamResponse("test-id", "gpt-4o", "Hello", stringPtr("stop"))

	if response.ID != "test-id" {
		t.Errorf("ID = %v, want test-id", response.ID)
	}
	if response.Choices[0].Delta.Content != "Hello" {
		t.Errorf("Content = %v, want Hello", response.Choices[0].Delta.Content)
	}
	if response.Choices[0].FinishReason == nil || *response.Choices[0].FinishReason != "stop" {
		t.Errorf("FinishReason = %v, want stop", response.Choices[0].FinishReason)
	}
}

func TestNewErrorResponse(t *testing.T) {
	response := NewErrorResponse("Test error", "test_error", "error_code")

	if response.Error.Message != "Test error" {
		t.Errorf("Message = %v, want Test error", response.Error.Message)
	}
	if response.Error.Type != "test_error" {
		t.Errorf("Type = %v, want test_error", response.Error.Type)
	}
	if response.Error.Code != "error_code" {
		t.Errorf("Code = %v, want error_code", response.Error.Code)
	}
}

// Helper function
func stringPtr(s string) *string {
	return &s
}
