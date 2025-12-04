package escriptorium

import "net/http"

type Client struct {
	username string
	password string
	basePath string

	token      string
	httpClient *http.Client
}

func NewClient(username, password, basePath string) *Client {
	return &Client{
		username:   username,
		password:   password,
		basePath:   basePath,
		httpClient: &http.Client{},
	}
}

func (c *Client) Authenticate() error {
	token, err := getAuthToken(c.username, c.password, c.basePath)
	if err != nil {
		return err
	}
	c.token = token
	return nil
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	if c.token == "" {
		if err := c.Authenticate(); err != nil {
			return nil, err
		}
	}
	req.Header.Set("Authorization", "Token "+c.token)

	return c.httpClient.Do(req)
}
