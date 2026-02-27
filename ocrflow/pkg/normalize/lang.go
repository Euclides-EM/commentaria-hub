package normalize

import "regexp"

var rulesLanguage = []rule{
	{regexp.MustCompile(`latin|latina|latino|latine|latein|latijn|latinum|latinit|la tine|latijnsche`), "Latin"},
	{regexp.MustCompile(`greek|graec|græc|grec|griech`), "Greek"},
	{regexp.MustCompile(`fran[çc]ois|francois|french`), "French"},
	{regexp.MustCompile(`italien|italian|italiana|thoscana|toscana`), "Italian"},
	{regexp.MustCompile(`spanish|espanol|española|traduzidas|castellano|hispanice`), "Spanish"},
	{regexp.MustCompile(`german|teutsch|teutscher|deutsch`), "German"},
	{regexp.MustCompile(`nederduyts|nederduytse|neerduid|neerduyts|neerdvyt|niderland`), "Dutch"},
	{regexp.MustCompile(`arabic`), "Arabic"},
	{regexp.MustCompile(`english|englishe`), "English"},
	{regexp.MustCompile(`romance|vulgar|volgar|vvlgare|vernacul|en nostre langve`), "General-Vernacular"},
}

func Language(lang string) string {
	return byRegex(rulesLanguage, "Other")(lang)
}
