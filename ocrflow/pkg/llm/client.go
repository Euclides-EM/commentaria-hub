package llm

import "errors"

type Client struct {
	openAIKey string
}

func (c *Client) Exec(prompt string, attachmentPath string) (map[string][]string, error) {
	return nil, errors.New("LLM client not implemented")
}

func NewClient(openAIKey string) *Client {
	return &Client{openAIKey: openAIKey}
}
