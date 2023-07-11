package main

import (
	_ "github.com/lib/pq"
	"russian_losses/pkg/api"
)

func main() {
	api.HandleRequests()
}
