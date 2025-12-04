package escriptorium

import "time"

type Project struct {
	ID               int           `json:"id"`
	Owner            string        `json:"owner"`
	Slug             string        `json:"slug"`
	DocumentsCount   int           `json:"documents_count"`
	SharedWithUsers  []interface{} `json:"shared_with_users"`
	SharedWithGroups []interface{} `json:"shared_with_groups"`
	Name             string        `json:"name"`
	Guidelines       interface{}   `json:"guidelines"`
	CreatedAt        time.Time     `json:"created_at"`
	UpdatedAt        time.Time     `json:"updated_at"`
	Tags             []interface{} `json:"tags"`
}

func (c *Client) GetProjects() ([]*Project, error) {
	return get[*Project](c, "api/projects/")
}
