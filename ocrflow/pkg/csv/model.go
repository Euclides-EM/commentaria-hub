package csv

type OptionOnDuplicate int

const (
	OptionOnDuplicateError OptionOnDuplicate = iota
	OptionOnDuplicateOverwrite
	OptionOnDuplicateIgnore
)

type Options struct {
	OnDuplicate OptionOnDuplicate
}

func DefaultOptions() Options {
	return Options{
		OnDuplicate: OptionOnDuplicateOverwrite,
	}
}

func IgnoreDuplicatesOptions() Options {
	return Options{
		OnDuplicate: OptionOnDuplicateIgnore,
	}
}
