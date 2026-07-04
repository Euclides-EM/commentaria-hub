module github.com/MiaMish/elements-dh/indexextractor

go 1.25

require (
	github.com/MiaMish/elements-dh/ocrflow v0.0.0-20260704123851-09b621b0c505
	github.com/avast/retry-go v3.0.0+incompatible
	github.com/joho/godotenv v1.5.1
	github.com/openai/openai-go/v3 v3.15.0
	github.com/samber/lo v1.52.0
)

replace github.com/MiaMish/elements-dh/ocrflow => github.com/Euclides-EM/commentaria-hub/ocrflow v0.0.0-20260704123851-09b621b0c505

require (
	github.com/stretchr/testify v1.11.1 // indirect
	github.com/tidwall/gjson v1.18.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/tidwall/sjson v1.2.5 // indirect
	golang.org/x/text v0.22.0 // indirect
)
