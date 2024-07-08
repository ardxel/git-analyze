package core

type LanguageEntry struct {
	Blank    int64 `json:"blank"`
	Comments int64 `json:"comments"`
	Lines    int64 `json:"lines"`
	Files    int64 `json:"files"`
}

func newLangEntry() *LanguageEntry {
	return &LanguageEntry{}
}
