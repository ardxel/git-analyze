package analyzer

import (
	"reflect"
	"testing"
)

func TestGetLang(t *testing.T) {
	tests := []struct {
		ext  string
		want string
	}{
		{"java", "Java"},
		{"py", "Python"},
		{"js", "JavaScript"},
		{"go", "Go"},
		{"cpp", "C++"},
	}

	for _, tt := range tests {
		t.Run(tt.ext, func(t *testing.T) {
			got := registry.GetLangByExt(tt.ext)
			if got != tt.want {
				t.Errorf("GetLang(%q) = %q; want %q", tt.ext, got, tt.want)
			}
		})
	}
}

func TestGetLineMarkers(t *testing.T) {
	tests := []struct {
		lang string
		want []string
	}{
		{"Python", []string{"#"}},
		{"Java", []string{"//"}},
		{"Lua", []string{"--"}},
	}

	for _, tt := range tests {
		t.Run(tt.lang, func(t *testing.T) {
			got := registry.GetLineComments(tt.lang)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetLineMarkers(%q) = %q; want %q", tt.lang, got, tt.want)
			}
		})
	}
}

func TestGetBlockMarkers(t *testing.T) {
	tests := []struct {
		lang string
		want [][]string
	}{
		{"Python", [][]string{{"\"\"\"", "\"\"\""}}},
		{"Java", [][]string{{"/*", "*/"}}},
		{"HTML", [][]string{{"<!--", "-->"}}},
		{"Haskell", [][]string{{"{-", "-}"}}},
		{"Ruby", [][]string{{":=begin", ":=end"}}},
	}

	for _, tt := range tests {
		t.Run(tt.lang, func(t *testing.T) {
			got := registry.GetBlockComments(tt.lang)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetBlockMarkers(%q) = %q; want %q", tt.lang, got, tt.want)
			}
		})
	}
}
