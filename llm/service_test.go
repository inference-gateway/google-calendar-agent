package llm_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/inference-gateway/google-calendar-agent/llm"
	"github.com/inference-gateway/google-calendar-agent/llm/mocks"
)

func TestService_ProcessNaturalLanguage_WithCounterfeiterMock(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		mockSetup      func(*mocks.FakeService)
		expectedResult *llm.ProcessingResult
		expectedError  string
	}{
		{
			name:  "successful event creation",
			input: "Create a meeting tomorrow at 2pm",
			mockSetup: func(mockService *mocks.FakeService) {
				mockService.ProcessNaturalLanguageReturns(&llm.ProcessingResult{
					Intent:     "create_event",
					Confidence: 0.95,
					Parameters: map[string]interface{}{
						"title":      "meeting",
						"start_time": "2pm",
						"date":       "tomorrow",
					},
					Response:       "I'll create a meeting for tomorrow at 2pm.",
					ProcessingTime: 150 * time.Millisecond,
				}, nil)
				mockService.IsEnabledReturns(true)
				mockService.GetProviderReturns("openai")
				mockService.GetModelReturns("gpt-4")
			},
			expectedResult: &llm.ProcessingResult{
				Intent:     "create_event",
				Confidence: 0.95,
				Parameters: map[string]interface{}{
					"title":      "meeting",
					"start_time": "2pm",
					"date":       "tomorrow",
				},
				Response:       "I'll create a meeting for tomorrow at 2pm.",
				ProcessingTime: 150 * time.Millisecond,
			},
		},
		{
			name:  "list events request",
			input: "Show me my events for next week",
			mockSetup: func(mockService *mocks.FakeService) {
				mockService.ProcessNaturalLanguageReturns(&llm.ProcessingResult{
					Intent:     "list_events",
					Confidence: 0.90,
					Parameters: map[string]interface{}{
						"time_range": "next week",
					},
					Response:       "Here are your events for next week:",
					ProcessingTime: 120 * time.Millisecond,
				}, nil)
				mockService.IsEnabledReturns(true)
				mockService.GetProviderReturns("anthropic")
				mockService.GetModelReturns("claude-3-sonnet")
			},
			expectedResult: &llm.ProcessingResult{
				Intent:     "list_events",
				Confidence: 0.90,
				Parameters: map[string]interface{}{
					"time_range": "next week",
				},
				Response:       "Here are your events for next week:",
				ProcessingTime: 120 * time.Millisecond,
			},
		},
		{
			name:  "service disabled",
			input: "Create a meeting",
			mockSetup: func(mockService *mocks.FakeService) {
				mockService.ProcessNaturalLanguageReturns(nil, assert.AnError)
				mockService.IsEnabledReturns(false)
				mockService.GetProviderReturns("")
				mockService.GetModelReturns("")
			},
			expectedError: assert.AnError.Error(),
		},
		{
			name:  "LLM asks clarifying question",
			input: "Schedule a meeting",
			mockSetup: func(mockService *mocks.FakeService) {
				mockService.ProcessNaturalLanguageReturns(&llm.ProcessingResult{
					Intent:         "clarification",
					Confidence:     0.9,
					Parameters:     map[string]interface{}{},
					Response:       "I'd be happy to help you schedule a meeting! Could you please provide more details like the date, time, and who should attend?",
					ProcessingTime: 100 * time.Millisecond,
				}, nil)
				mockService.IsEnabledReturns(true)
				mockService.GetProviderReturns("openai")
				mockService.GetModelReturns("gpt-4")
			},
			expectedResult: &llm.ProcessingResult{
				Intent:         "clarification",
				Confidence:     0.9,
				Parameters:     map[string]interface{}{},
				Response:       "I'd be happy to help you schedule a meeting! Could you please provide more details like the date, time, and who should attend?",
				ProcessingTime: 100 * time.Millisecond,
			},
		},
		{
			name:  "LLM provides informational response",
			input: "What's the weather like?",
			mockSetup: func(mockService *mocks.FakeService) {
				mockService.ProcessNaturalLanguageReturns(&llm.ProcessingResult{
					Intent:         "question",
					Confidence:     0.8,
					Parameters:     map[string]interface{}{},
					Response:       "I'm a calendar assistant and can help you manage your calendar events. I don't have access to weather information, but I can help you schedule events, list your appointments, or check your availability.",
					ProcessingTime: 80 * time.Millisecond,
				}, nil)
				mockService.IsEnabledReturns(true)
				mockService.GetProviderReturns("openai")
				mockService.GetModelReturns("gpt-4")
			},
			expectedResult: &llm.ProcessingResult{
				Intent:         "question",
				Confidence:     0.8,
				Parameters:     map[string]interface{}{},
				Response:       "I'm a calendar assistant and can help you manage your calendar events. I don't have access to weather information, but I can help you schedule events, list your appointments, or check your availability.",
				ProcessingTime: 80 * time.Millisecond,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mocks.FakeService{}
			tt.mockSetup(mockService)

			ctx := context.Background()
			result, err := mockService.ProcessNaturalLanguage(ctx, tt.input)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}

			assert.Equal(t, 1, mockService.ProcessNaturalLanguageCallCount())
			receivedCtx, receivedInput := mockService.ProcessNaturalLanguageArgsForCall(0)
			assert.Equal(t, ctx, receivedCtx)
			assert.Equal(t, tt.input, receivedInput)
		})
	}
}

func TestService_ConfigMethods_WithCounterfeiterMock(t *testing.T) {
	mockService := &mocks.FakeService{}

	mockService.IsEnabledReturns(true)
	mockService.GetProviderReturns("openai")
	mockService.GetModelReturns("gpt-4-turbo")

	enabled := mockService.IsEnabled()
	assert.True(t, enabled)
	assert.Equal(t, 1, mockService.IsEnabledCallCount())

	provider := mockService.GetProvider()
	assert.Equal(t, "openai", provider)
	assert.Equal(t, 1, mockService.GetProviderCallCount())

	model := mockService.GetModel()
	assert.Equal(t, "gpt-4-turbo", model)
	assert.Equal(t, 1, mockService.GetModelCallCount())
}

func TestService_MultipleCallsTracking_WithCounterfeiterMock(t *testing.T) {
	mockService := &mocks.FakeService{}

	mockService.ProcessNaturalLanguageReturnsOnCall(0, &llm.ProcessingResult{Intent: "create_event"}, nil)
	mockService.ProcessNaturalLanguageReturnsOnCall(1, &llm.ProcessingResult{Intent: "list_events"}, nil)
	mockService.ProcessNaturalLanguageReturnsOnCall(2, nil, assert.AnError)

	ctx := context.Background()

	result1, err1 := mockService.ProcessNaturalLanguage(ctx, "create meeting")
	assert.NoError(t, err1)
	assert.Equal(t, "create_event", result1.Intent)

	result2, err2 := mockService.ProcessNaturalLanguage(ctx, "list events")
	assert.NoError(t, err2)
	assert.Equal(t, "list_events", result2.Intent)

	result3, err3 := mockService.ProcessNaturalLanguage(ctx, "invalid input")
	assert.Error(t, err3)
	assert.Nil(t, result3)

	assert.Equal(t, 3, mockService.ProcessNaturalLanguageCallCount())

	_, input1 := mockService.ProcessNaturalLanguageArgsForCall(0)
	assert.Equal(t, "create meeting", input1)

	_, input2 := mockService.ProcessNaturalLanguageArgsForCall(1)
	assert.Equal(t, "list events", input2)

	_, input3 := mockService.ProcessNaturalLanguageArgsForCall(2)
	assert.Equal(t, "invalid input", input3)
}
