package main

import (
	"russian_losses/pkg/api"

	_ "github.com/lib/pq"
)

func main() {
	api.HandleRequests()
}
