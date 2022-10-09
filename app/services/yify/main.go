package main

import (
	"fmt"
	"os"

	"github.com/ezratameno/yify/internal/yify"
)

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run() error {
	client, err := yify.New()
	if err != nil {
		return err
	}
	client.CollectMovies()
	return nil
}
