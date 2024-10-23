package api

import (
	"context"
	"fmt"
	"git-analyzer/pkg/analyzer"
	"git-analyzer/pkg/config"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v63/github"
)

type AnalyzeResultMap struct {
	RepoSizeLimit int64                `redis:"repo_size_limit"`
	IsProd        bool                 `redis:"is_prod"`
	ParallelMode  bool                 `redis:"parallel_mode"`
	Languages     []*analyzer.Language `redis:"languages"`
	TotalLines    int32                `redis:"total_lines"`
	TotalFiles    int32                `redis:"total_files"`
	TotalBlank    int32                `redis:"total_blank"`
	TotalComments int32                `redis:"total_comments"`
	FetchSpeed    time.Duration        `redis:"fetch_speed"`
	AnalysisSpeed time.Duration        `redis:"analysis_speed"`
	Error         string               `redis:"error"`
}

/*
 * FUNC UTILS FOR TEMPLATES
 */

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

func badgeURL(langname string) string {
	badgeURL, ok := badges[langname]

	if ok {
		name, logo, color := badgeURL.Name, badgeURL.Logo, badgeURL.Color
		return fmt.Sprintf("https://img.shields.io/badge/%s?logo=%s&logoColor=%s", name, logo, color)
	}

	langname = strings.ReplaceAll(langname, " ", "%20")
	return fmt.Sprintf("https://img.shields.io/badge/%s-000000?logo=github&logoColor=fff", langname)

}

func formatTime(d time.Duration) string {

	totalSeconds := int(d.Seconds())
	minutes := totalSeconds / 60
	seconds := totalSeconds % 60
	milliseconds := d.Milliseconds() % 1000

	if minutes > 0 {
		return fmt.Sprintf("%02d.%02d m", minutes, seconds)
	}

	return fmt.Sprintf("%02d.%03d s", seconds, milliseconds)
}

func validateRepoURL(rawurl string) bool {
	parsedURL, err := url.Parse(rawurl)

	if err != nil {
		return false
	}

	if parsedURL.Host != "github.com" {
		return false
	}

	matches := githubRegexp.FindStringSubmatch(parsedURL.Path)

	if matches == nil || len(matches) < 3 {
		return false
	}

	return true
}

func extractMeta(rawurl string) (onwer string, repo string, err error) {
	parsedURL, err := url.Parse(rawurl)

	if err != nil {
		return "", "", fmt.Errorf("invalid url: %s", err)
	}

	if parsedURL.Host != "github.com" {
		if parsedURL.Host == "" {
			parsedURL.Host = "none"
		}
		return "", "", fmt.Errorf("Wrong host, expected github.com, got %s", parsedURL.Host)
	}

	matches := githubRegexp.FindStringSubmatch(parsedURL.Path)

	if matches == nil || len(matches) < 3 {
		return "", "", fmt.Errorf("Url must contain owner and repo: %s", rawurl)
	}

	owner, repoName := matches[1], matches[2]

	return owner, repoName, nil
}

func RepoIsExists(owner, name string) bool {
	ctx := context.Background()

	_, _, err := githubClient.Repositories.Get(ctx, owner, name)

	if _, ok := getRateLimitError(err); ok {
		return true
	}

	fmt.Println("ERROR IS EXISTS: ", err)

	return err == nil
}

// returns repo size in bytes
func fetchRepoSize(owner, name string) (int64, error) {

	ctx := context.Background()
	repo, res, err := githubClient.Repositories.Get(ctx, owner, name)

	if errRateLimit, ok := getRateLimitError(err); ok {
		return 0, errRateLimit
	}

	if res.StatusCode == 404 {
		return 0, fmt.Errorf("Repository not found")
	}

	repoSize := int64(*repo.Size) * 1024 // bytes

	if config.Vars.Debug {
		log.Printf("Repo: %v, Size: %d MB\n", name, repoSize/1024/1024)
	}

	return repoSize, nil
}

func getRateLimitError(err error) (error, bool) {
	if err != nil {
		if rle, rateLimitOk := err.(*github.RateLimitError); rateLimitOk {
			return fmt.Errorf("GitHub API rate limit exceeded. Try again in %s", rle.Rate.Reset), true
		}
		if arle, abuseRateLimitOk := err.(*github.AbuseRateLimitError); abuseRateLimitOk {
			return fmt.Errorf("GitHub API rate limit exceeded. Try again in %s", arle.RetryAfter), true
		}
	}
	return nil, false
}

func getOwner(ctx *gin.Context) string {
	value, ok := ctx.Get("repo_owner")

	if !ok {
		return ""
	}

	return value.(string)
}

func getName(ctx *gin.Context) string {
	value, ok := ctx.Get("repo_name")

	if !ok {
		return ""
	}

	return value.(string)
}
