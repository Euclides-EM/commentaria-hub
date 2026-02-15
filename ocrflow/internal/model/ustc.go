package model

type USTC struct {
	USTCId        int      `json:"ustc_id"`
	Authors       []string `json:"authors"`
	ShortTitle    string   `json:"short_title"`
	Publishers    []string `json:"publishers"`
	City          *string  `json:"city"`
	Year          *int     `json:"year"`
	Languages     []string `json:"languages"`
	Digitizations []string `json:"digitizations"`
	Format        *string  `json:"format"`
}
