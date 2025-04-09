package summary

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/sashabaranov/go-openai"
)

type OpenAI struct {
	client  *openai.Client
	prompt  string
	model   string
	enabled bool
	mu      sync.Mutex
}

func NewOpenAI(apiKey, model, prompt string) *OpenAI {
	s := &OpenAI{
		client: openai.NewClient(apiKey),
		prompt: prompt,
		model:  model,
	}

	log.Printf("openai summarizer is enabled: %v", apiKey != "")

	if apiKey != "" {
		s.enabled = true
	}

	return s
}

func (s *OpenAI) GetClient() *openai.Client {
	return s.client
}

func (s *OpenAI) Summarize(text string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.enabled {
		return "", fmt.Errorf("openai summarizer is disabled")
	}

	request := openai.ChatCompletionRequest{
		Model: s.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: s.prompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: text,
			},
		},
		MaxTokens:   1024,
		Temperature: 1,
		TopP:        1,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	resp, err := s.client.CreateChatCompletion(ctx, request)
	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", errors.New("no choices in openai response")
	}

	rawSummary := strings.TrimSpace(resp.Choices[0].Message.Content)
	if strings.HasSuffix(rawSummary, ".") {
		return rawSummary, nil
	}

	// cut all after the last ".":
	sentences := strings.Split(rawSummary, ".")

	return strings.Join(sentences[:len(sentences)-1], ".") + ".", nil
}

func (s *OpenAI) SetCaption(prompt string, image string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.enabled {
		return "", fmt.Errorf("openai summarizer is disabled")
	}

	imgUrl := openai.ChatMessageImageURL{
		URL: image,
	}

	contentImg := openai.ChatMessagePart{
		ImageURL: &imgUrl,
		Type:     openai.ChatMessagePartTypeImageURL,
	}

	contentText := openai.ChatMessagePart{
		Text: prompt,
		Type: openai.ChatMessagePartTypeText,
	}

	contentSystem := openai.ChatMessagePart{
		Text: "Придумай на русском языке прикольное название для картинки мема и добавь emoji",
		Type: openai.ChatMessagePartTypeText,
	}

	// Создаём JSON-объект в виде структуры
	data := []openai.ChatCompletionMessage{
		{
			Role:         "user",
			MultiContent: []openai.ChatMessagePart{contentImg, contentText},
		},
		{
			Role:         "system",
			MultiContent: []openai.ChatMessagePart{contentSystem},
		},
	}

	request := openai.ChatCompletionRequest{
		Model:       s.model,
		Messages:    data,
		MaxTokens:   1024,
		Temperature: 1,
		TopP:        1,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	resp, err := s.client.CreateChatCompletion(ctx, request)
	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", errors.New("no choices in openai response")
	}

	rawSummary := strings.TrimSpace(resp.Choices[0].Message.Content)

	return rawSummary, nil
}
