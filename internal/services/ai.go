package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"memology-backend/internal/config"
)

type AIService interface {
	GenerateMeme(ctx context.Context, userInput, style string) (string, error)
	GenerateTemplateMeme(ctx context.Context, req GenerateTemplateRequest) (*GenerateTemplateResponse, error)
	GetTaskStatus(ctx context.Context, taskID string) (*TaskStatusResponse, error)
	GetTaskResult(ctx context.Context, taskID string) ([]byte, error)
	GetAvailableStyles(ctx context.Context) ([]string, error)
}

type aiService struct {
	config *config.AIConfig
	client *http.Client
}

func NewAIService(cfg *config.AIConfig) AIService {
	return &aiService{
		config: cfg,
		client: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}

type GenerateMemeRequest struct {
	UserInput string `json:"user_input"`
	Style     string `json:"style,omitempty"`
}

type GenerateMemeResponse struct {
	TaskID string `json:"task_id"`
}

type TaskStatusResponse struct {
	Status     string `json:"status"`
	TaskID     string `json:"task_id"`
	ResultPath string `json:"result_path,omitempty"`
}

// GenerateTemplateRequest - запрос на генерацию шаблонного мема через memegen.link
type GenerateTemplateRequest struct {
	Context string `json:"context" validate:"required" example:"Кот пьет кофе"`
	Width   int    `json:"width,omitempty" example:"512"`
	Height  int    `json:"height,omitempty" example:"512"`
}

// GenerateTemplateResponse - ответ с готовым шаблонным мемом
type GenerateTemplateResponse struct {
	URL      string          `json:"url"`
	Template string          `json:"template"`
	Text     json.RawMessage `json:"text"` // может быть строкой или массивом
}

// GetTextStrings возвращает текст как массив строк
func (r *GenerateTemplateResponse) GetTextStrings() []string {
	// Попробуем как массив
	var arr []string
	if err := json.Unmarshal(r.Text, &arr); err == nil {
		return arr
	}
	// Попробуем как строку
	var str string
	if err := json.Unmarshal(r.Text, &str); err == nil {
		return []string{str}
	}
	return nil
}

func (s *aiService) GenerateMeme(ctx context.Context, userInput, style string) (string, error) {
	reqBody := GenerateMemeRequest{
		UserInput: userInput,
		Style:     style,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.config.BaseURL+"/api/memes/generate", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("AI service returned status %d: %s", resp.StatusCode, string(body))
	}

	var result GenerateMemeResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.TaskID, nil
}

func (s *aiService) GetTaskStatus(ctx context.Context, taskID string) (*TaskStatusResponse, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api/memes/task/%s", s.config.BaseURL, taskID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("AI service returned status %d: %s", resp.StatusCode, string(body))
	}

	var result TaskStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func (s *aiService) GetTaskResult(ctx context.Context, taskID string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api/memes/task/%s/result", s.config.BaseURL, taskID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("AI service returned status %d: %s", resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
}

type StyleObject struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (s *aiService) GetAvailableStyles(ctx context.Context) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", s.config.BaseURL+"/api/memes/styles", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("AI service returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var stylesArray []string
	if err := json.Unmarshal(body, &stylesArray); err == nil {
		return stylesArray, nil
	}

	var stylesObjectArray []StyleObject
	if err := json.Unmarshal(body, &stylesObjectArray); err == nil {
		result := make([]string, 0, len(stylesObjectArray))
		for _, style := range stylesObjectArray {
			result = append(result, style.Name)
		}
		return result, nil
	}

	var stylesObject map[string]interface{}
	if err := json.Unmarshal(body, &stylesObject); err == nil {
		if styles, ok := stylesObject["styles"].([]interface{}); ok {
			result := make([]string, 0, len(styles))
			for _, style := range styles {
				if s, ok := style.(string); ok {
					result = append(result, s)
				}
			}
			return result, nil
		}
	}

	return nil, fmt.Errorf("failed to parse styles response: %s", string(body))
}

// GenerateTemplateMeme генерирует шаблонный мем через memegen.link API (синхронно)
func (s *aiService) GenerateTemplateMeme(ctx context.Context, req GenerateTemplateRequest) (*GenerateTemplateResponse, error) {
	// Устанавливаем размеры по умолчанию, если не указаны
	if req.Width == 0 {
		req.Width = 512
	}
	if req.Height == 0 {
		req.Height = 512
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", s.config.BaseURL+"/api/memes/generate-template", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("AI service returned status %d: %s", resp.StatusCode, string(body))
	}

	var result GenerateTemplateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}
