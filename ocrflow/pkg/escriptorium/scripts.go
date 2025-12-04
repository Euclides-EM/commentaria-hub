package escriptorium

type Script struct {
	Id            int    `json:"id"`
	Name          string `json:"name"`
	NameFr        string `json:"name_fr"`
	IsoCode       string `json:"iso_code"`
	TextDirection string `json:"text_direction"`
	BlankChar     string `json:"blank_char"`
}

func (c *Client) GetScripts() ([]*Script, error) {
	return get[*Script](c, "api/scripts")
}
