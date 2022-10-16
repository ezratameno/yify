package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"

	fyne "fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/ezratameno/yify/business/data/store/movie"
)

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run() error {
	// ==================================================
	// get the movies
	host := flag.String("host", "http://localhost:3000", "host of the api")
	flag.Parse()
	resp, err := http.Get(fmt.Sprintf("%s/v1/movies", *host))
	if err != nil {
		return err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var movies []movie.Movie
	err = json.Unmarshal(body, &movies)
	if err != nil {
		return err
	}
	sort.SliceStable(movies, func(i, j int) bool {
		y1, _ := strconv.Atoi(movies[i].Year)
		y2, _ := strconv.Atoi(movies[j].Year)

		return y1 > y2
	})

	// ===========================================================
	// Display the movies

	a := app.New()
	a.Settings().SetTheme(theme.DarkTheme())
	w := a.NewWindow("Ezra's Torrents")
	// fix starting window size
	w.Resize(fyne.NewSize(1200, 800))

	listView := widget.NewList(
		// return the length of the list
		func() int {
			return len(movies)
			// create the item that the list is gonna render
		}, func() fyne.CanvasObject {
			return widget.NewLabel("template")
			// update the rendering of the object
		}, func(id widget.ListItemID, object fyne.CanvasObject) {
			object.(*widget.Label).SetText(fmt.Sprintf("%s (%s)", movies[id].Name, movies[id].Year))
		})

	// right side container
	containerDisplay := container.NewVSplit(container.NewMax(), container.NewMax())
	contentContainer := container.NewMax(containerDisplay)
	listView.OnSelected = func(id widget.ListItemID) {

		contentContainer.Objects = nil

		btnContainer := container.NewHBox()

		for quality, link := range movies[id].DownloadLinks {
			btn := widget.NewButton(quality, func() {})
			btn.OnTapped = func() {
				Download(movies[id], link)
			}
			btn.Position()
			btnContainer.Add(btn)
		}

		// create image
		img, _ := loadResourceFromURLString(movies[id].ImageUrl)
		image := canvas.NewImageFromResource(img)
		image.FillMode = canvas.ImageFillContain
		infoContainer := container.NewVSplit(image, displayInfo(movies[id]))
		// fix initail size
		infoContainer.Offset = 1

		containerDisplay := container.NewVSplit(infoContainer, btnContainer)
		containerDisplay.Offset = 1
		contentContainer.Add(containerDisplay)
		contentContainer.Refresh()
	}

	split := container.NewHSplit(
		listView,
		contentContainer,
	)
	split.Offset = 0.85

	w.SetContent(split)

	w.ShowAndRun()

	for _, movie := range movies {
		fmt.Printf("%+v\n", movie)
	}

	return nil
}

func displayInfo(movie movie.Movie) *widget.Label {
	c := widget.NewLabel(
		fmt.Sprintf(`Movie page: %s
Name: %s
Description: %s
Year: %s
Categories: %s`, movie.PageUrl, movie.Name, movie.Description, movie.Year, strings.Join(movie.Categories, ",")))
	c.Wrapping = fyne.TextWrapWord
	return c
}
