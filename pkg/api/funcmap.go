package api

import (
	"fmt"
	"strings"
	"time"
)

var badges = map[string]*struct {
	Name  string
	Logo  string
	Color string
}{
	"Bourne Shell": {"Bash-4EAA25", "gnubash", "fff"},
	"C":            {"C-00599C", "c", "white"},
	"C++":          {"C++-%2300599C", "c%2B%2B", "white"},
	"C#":           {"C%23-%23239120", "cshrp", "white"},
	"CoffeeScript": {"CoffeeScript-2F2625", "coffeescript", "fff"},
	"Clojure":      {"Clojure-5881D8", "clojure", "fff"},
	"Crystal":      {"Crystal-000", "crystal", "fff"},
	"CSS":          {"CSS-1572B6", "css3", "fff"},
	"Dart":         {"Dart-%230175C2", "dart", "white"},
	"Elixir":       {"Elixir-%234B275F", "elixir", "white"},
	"Elm":          {"Elm-1293D8", "elm", "fff"},
	"Erlang":       {"Erlang-A90533", "erlang", "fff"},
	"F#":           {"F%23-378BBA", "fsharp", "fff"},
	"Flutter":      {"Flutter-02569B", "flutter", "fff"},
	"Fortran":      {"Fortran-734F96", "fortran", "fff"},
	"Go":           {"Go-%2300ADD8", "go", "white"},
	"Haskell":      {"Haskell-5e5086", "haskell", "white"},
	"HTML":         {"HTML-%23E34F26", "html5", "white"},
	"HTMX":         {"HTMX-36C", "htmx", "fff"},
	"Java":         {"Java-%23ED8B00", "openjdk", "white"},
	"JavaScript":   {"JavaScript-F7DF1E", "javascript", "000"},
	"JSON":         {"JSON-000", "json", "fff"},
	"Kotlin":       {"Kotlin-%237F52FF", "kotlin", "white"},
	"Lua":          {"Lua-%232C2D72", "lua", "white"},
	"Markdown":     {"Markdown-%23000000", "markdown", "white"},
	"MDX":          {"MDX-1B1F24", "mdx", "fff"},
	"Nim":          {"Nim-%23FFE953", "nim", "white"},
	"Nix":          {"Nix-5277C3", "NixOS", "white"},
	"OCaml":        {"OCaml-EC6813", "ocaml", "fff"},
	"Odin":         {"Odin-1E5184", "odinlang", "fff"},
	"Objective-C":  {"Objective--C-%233A95E3", "apple", "white"},
	"Perl":         {"Perl-%2339457E", "perl", "white"},
	"PHP":          {"php-%23777BB4", "php", "white"},
	"Python":       {"Python-3776AB", "python", "fff"},
	"R":            {"R-%23276DC3", "r", "white"},
	"Ruby":         {"Ruby-%23CC342D", "ruby", "white"},
	"Rust":         {"Rust-%23000000", "rust", "white"},
	"Sass":         {"Sass-C69", "sass", "fff"},
	"Scratch":      {"Scratch-4D97FF", "scratch", "fff"},
	"Scala":        {"Scala-%23DC322F", "scala", "white"},
	"Solidity":     {"Solidity-363636", "solidity", "fff"},
	"Swift":        {"Swift-F54A2A", "swift", "white"},
	"TypeScript":   {"TypeScript-3178C6", "typescript", "fff"},
	"V":            {"V-5D87BF", "v", "fff"},
	"WebAssembly":  {"WebAssembly-654FF0", "webassembly", "fff"},
	"YAML":         {"YAML-CB171E", "yaml", "fff"},
	"Zig":          {"Zig-F7A41D", "zig", "fff"},
}

func BadgeURL(langname string) string {
	badgeURL, ok := badges[langname]

	if ok {
		name, logo, color := badgeURL.Name, badgeURL.Logo, badgeURL.Color
		return fmt.Sprintf("https://img.shields.io/badge/%s?logo=%s&logoColor=%s", name, logo, color)
	}

	langname = strings.ReplaceAll(langname, " ", "%20")
	return fmt.Sprintf("https://img.shields.io/badge/%s-000000?logo=github&logoColor=fff", langname)

}

func FormatTime(d time.Duration) string {

	totalSeconds := int(d.Seconds())
	minutes := totalSeconds / 60
	seconds := totalSeconds % 60
	milliseconds := d.Milliseconds() % 1000

	if minutes > 0 {
		return fmt.Sprintf("%02d.%02d m", minutes, seconds)
	}

	return fmt.Sprintf("%02d.%03d s", seconds, milliseconds)
}
