package main

import "git-analyzer/pkg/api"

func main() {
	api := api.New()
	api.Start()
}
