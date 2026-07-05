module github.com/MiaMish/elements-dh/correspondence_ingest

go 1.25

require (
	github.com/Euclides-EM/commentaria-hub/ocrflow v0.0.0
	github.com/joho/godotenv v1.5.1
)

replace github.com/Euclides-EM/commentaria-hub/ocrflow => ../ocrflow

require (
	github.com/avast/retry-go v3.0.0+incompatible // indirect
	github.com/openai/openai-go/v3 v3.15.0 // indirect
	github.com/samber/lo v1.52.0 // indirect
	github.com/tidwall/gjson v1.18.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/tidwall/sjson v1.2.5 // indirect
	golang.org/x/text v0.31.0 // indirect
)
