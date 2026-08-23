package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type AIService struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
}

type ChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func NewAIService() *AIService {
	return &AIService{
		apiKey:  os.Getenv("OPENROUTER_API_KEY"),
		baseURL: "https://openrouter.ai/api/v1",
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// Chat sends a chat message to OpenRouter
func (s *AIService) Chat(messages []ChatMessage) (string, error) {
	if s.apiKey == "" {
		return "", fmt.Errorf("OPENROUTER_API_KEY not set")
	}

	request := ChatRequest{
		Model:    "openrouter/auto",
		Messages: messages,
		Stream:   false,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", s.baseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("HTTP-Referer", "https://communityshield.org")
	req.Header.Set("X-Title", "CommunityShield")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", err
	}

	if chatResp.Error != nil {
		return "", fmt.Errorf("API error: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no response from AI")
	}

	return chatResp.Choices[0].Message.Content, nil
}

// Chatbot - AI-powered assistant for citizens
func (s *AIService) Chatbot(question, userRole string) (string, error) {
	messages := []ChatMessage{
		{
			Role: "system",
			Content: `You are CommunityShield AI, a helpful security assistant for Nigerian communities.
Your role is to:
1. Provide safety tips and security advice
2. Help users report incidents
3. Explain how the platform works
4. Give general security information
5. Be friendly and culturally aware

If you don't know something, be honest and suggest they contact their local security unit.`,
		},
		{
			Role:    "user",
			Content: fmt.Sprintf("User role: %s\nQuestion: %s", userRole, question),
		},
	}

	return s.Chat(messages)
}

// AnalyzeImage - AI image analysis for crime scene photos
func (s *AIService) AnalyzeImage(imageDescription string) (string, error) {
	messages := []ChatMessage{
		{
			Role: "system",
			Content: `You are a forensic image analyst for CommunityShield.
Analyze the image description and provide:
1. Key observations
2. Potential evidence identification
3. Safety implications
4. Recommended actions`,
		},
		{
			Role:    "user",
			Content: fmt.Sprintf("Image description: %s", imageDescription),
		},
	}

	return s.Chat(messages)
}

// AnalyzeLocationRisk analyzes security risk for a specific location
func (s *AIService) AnalyzeLocationRisk(latitude, longitude float64, locationName string, recentIncidents string) (string, error) {
	messages := []ChatMessage{
		{
			Role: "system",
			Content: `You are a security intelligence analyst for CommunityShield in Nigeria.
Analyze location security risks and provide actionable insights.
Consider: local context, recent incidents, and practical safety measures.`,
		},
		{
			Role: "user",
			Content: fmt.Sprintf(`Analyze security risk for this location:
Location: %s (Lat: %f, Lng: %f)
Recent Incidents: %s

Provide:
1. Risk Level (Low/Medium/High/Extreme)
2. Key Risk Factors
3. Safety Recommendations
4. Emergency Contacts (if known)`, locationName, latitude, longitude, recentIncidents),
		},
	}

	return s.Chat(messages)
}

// AnalyzeNewsSentiment analyzes news articles for security sentiment
func (s *AIService) AnalyzeNewsSentiment(newsContent string) (string, error) {
	messages := []ChatMessage{
		{
			Role: "system",
			Content: `You are a security news analyst for CommunityShield.
Analyze news content for security implications and sentiment.
Focus on: threat levels, affected areas, and community impact.`,
		},
		{
			Role: "user",
			Content: fmt.Sprintf(`Analyze this news for security implications:
%s

Provide:
1. Sentiment Score (Positive/Neutral/Negative)
2. Threat Level
3. Affected Locations
4. Key Risks Identified
5. Recommended Actions`, newsContent),
		},
	}

	return s.Chat(messages)
}

// GenerateSecurityWarning generates a security warning based on incidents
func (s *AIService) GenerateSecurityWarning(incidentsData string) (string, error) {
	messages := []ChatMessage{
		{
			Role: "system",
			Content: `You are a security warning system for CommunityShield.
Generate clear, actionable security warnings for communities.
Be specific, practical, and culturally appropriate for Nigeria.`,
		},
		{
			Role: "user",
			Content: fmt.Sprintf(`Based on these recent incidents, generate a security warning:
%s

Format:
[WARNING TYPE]
[Location/Area]
[Description]
[Recommended Actions]
[Contact Information]`, incidentsData),
		},
	}

	return s.Chat(messages)
}

// AnalyzeIncidentPatterns analyzes incident patterns by location
func (s *AIService) AnalyzeIncidentPatterns(location string, incidents []string) (string, error) {
	incidentsText := ""
	for i, inc := range incidents {
		incidentsText += fmt.Sprintf("%d. %s\n", i+1, inc)
	}

	messages := []ChatMessage{
		{
			Role: "system",
			Content: `You are a security pattern analyst for CommunityShield.
Identify crime patterns, hotspots, and emerging threats.
Provide data-driven insights for proactive security measures.`,
		},
		{
			Role: "user",
			Content: fmt.Sprintf(`Analyze incident patterns in %s:
%s

Provide:
1. Pattern Summary
2. Hotspots Identified
3. Time Patterns
4. Recommendations`, location, incidentsText),
		},
	}

	return s.Chat(messages)
}

// GetSmartSafetyTips provides context-aware safety tips
func (s *AIService) GetSmartSafetyTips(location, userRole, timeOfDay, recentThreats string) (string, error) {
	messages := []ChatMessage{
		{
			Role: "system",
			Content: `You are a safety advisor for CommunityShield.
Provide personalized, context-aware safety tips.
Consider: location, user role, time of day, and current threats.`,
		},
		{
			Role: "user",
			Content: fmt.Sprintf(`Provide safety tips for:
Location: %s
Role: %s
Time: %s
Recent Threats: %s

Give 5 specific, actionable tips.`, location, userRole, timeOfDay, recentThreats),
		},
	}

	return s.Chat(messages)
}

// PredictRiskHotspots predicts potential risk hotspots
func (s *AIService) PredictRiskHotspots(historicalData string) (string, error) {
	messages := []ChatMessage{
		{
			Role: "system",
			Content: `You are a predictive security analyst for CommunityShield.
Analyze historical data to predict potential risk hotspots.
Be specific about locations, timing, and types of risks.`,
		},
		{
			Role: "user",
			Content: fmt.Sprintf(`Based on this historical incident data, predict risk hotspots:
%s

Provide:
1. High-Risk Areas
2. Time Periods
3. Types of Incidents
4. Preventive Measures`, historicalData),
		},
	}

	return s.Chat(messages)
}

// SummarizeCase provides a concise case summary
func (s *AIService) SummarizeCase(caseData string) (string, error) {
	messages := []ChatMessage{
		{
			Role: "system",
			Content: `You are a security case summarizer. Create concise, informative summaries.
Focus on key facts, status, and critical action items.`,
		},
		{
			Role: "user",
			Content: fmt.Sprintf(`Summarize this case concisely:
%s`, caseData),
		},
	}

	return s.Chat(messages)
}
