package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	coreMovie "github.com/ezratameno/yify/business/core/movie"
	"github.com/ezratameno/yify/internal/yify"
	"github.com/sirupsen/logrus"

	"github.com/ezratameno/yify/business/data/schema"
	"github.com/ezratameno/yify/business/data/store/movie"
	"github.com/ezratameno/yify/business/sys/database"
)

func main() {
	err := migrate()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func migrate() error {
	cfg := database.Config{
		User:         "postgres",
		Password:     "postgres",
		Host:         "localhost",
		Name:         "postgres",
		MaxIdleConns: 0,
		MaxOpenConns: 0,
		DisableTLS:   true,
	}

	db, err := database.Open(cfg)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer db.Close()

	if err := schema.Migrate(db); err != nil {
		return fmt.Errorf("migrate database: %w", err)
	}
	fmt.Println("migrations complete")

	core := coreMovie.NewCore(logrus.New().WithField("service", "admin"), db)

	// ===================================
	// fill the db only in case we don't have any data
	movies, _ := core.GetMovies()
	if len(movies) != 0 {
		fmt.Println("Already have movies, not going to populate the db...")

		return nil
	}
	// ====================================
	// add movies data to the movies table
	client, err := yify.New()
	if err != nil {
		return err
	}
	movies = client.CollectMovies()

	for id, newMovie := range movies {
		// marshell the download links
		downloadLinks, err := json.Marshal(newMovie.DownloadLinks)
		if err != nil {
			fmt.Println(err)
			continue
		}
		nm := movie.NewMovie{
			ID:            id,
			Name:          newMovie.Name,
			Categories:    strings.Join(newMovie.Categories, "/"),
			ImageUrl:      newMovie.ImageUrl,
			PageUrl:       newMovie.PageUrl,
			Year:          newMovie.Year,
			Description:   newMovie.Description,
			DownloadLinks: string(downloadLinks),
		}
		nm, err = core.Create(nm)
		if err != nil {
			return err
		}
	}
	fmt.Println("seeding database completed")
	return nil
}
