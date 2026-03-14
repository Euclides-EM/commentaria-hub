package model

type USTC struct {
	USTCId        int      `json:"ustc_id"`
	AlreadyExists bool     `json:"already_exists" readonly:"true"`
	Authors       []string `json:"authors" readonly:"true"`
	ShortTitle    string   `json:"short_title" readonly:"true"`
	Publishers    []string `json:"publishers" readonly:"true"`
	City          *string  `json:"city" readonly:"true"`
	Year          *int     `json:"year" readonly:"true"`
	Languages     []string `json:"languages" readonly:"true"`
	Digitizations []string `json:"digitizations" readonly:"true"`
	Format        *string  `json:"format" readonly:"true"`
}
