package markdown

import "regexp"

var printerErrorCorrectionPattern = regexp.MustCompile(`([^\s{}]*)\{printer-error-correction:([^}]*)\}`)

func ResolvePrinterErrorCorrections(text string, applyCorrection bool) string {
	if applyCorrection {
		return printerErrorCorrectionPattern.ReplaceAllString(text, "${2}")
	}
	return printerErrorCorrectionPattern.ReplaceAllString(text, "${1}")
}
