package escriptorium

type PartType struct {
	PK   int    `json:"pk"`
	Name string `json:"name"`
}

func (c *Client) GetPartTypes() ([]*PartType, error) {
	return get[*PartType](c, "api/types/part/")
}
