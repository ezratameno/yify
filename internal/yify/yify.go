// Package yify provides support to getting movies from the yify website.
package yify

import (
	"fmt"
	"strings"
	"sync"

	"github.com/ezratameno/yify/business/data/store/movie"
	"github.com/gocolly/colly"
)

type Client struct {
	c *colly.Collector
}

const (
	Domain   = "yts.mx"
	HomePage = "https://yts.mx/browse-movies/0/all/all/0/year/0/all"
)

func New() (*Client, error) {
	c := colly.NewCollector(
		colly.AllowedDomains(Domain),
	)
	return &Client{
		c: c,
	}, nil
}

// CollectMovies scarp the yify site.
func (y *Client) CollectMovies() []movie.Movie {
	var wg sync.WaitGroup
	var movies []movie.Movie
	movieLinks, nextPage := y.getMovieLinkPerPage(HomePage)
	ch := make(chan movie.Movie)
	for {
		done := make(chan struct{})
		go y.CollectMovieDetails(ch, done, &wg, movieLinks)

		// go through all the links in the page.
	outer:
		for {
			select {
			case <-done:
				break outer
			case movie := <-ch:
				movies = append(movies, movie)
			}

		}

		if nextPage == "https://yts.mx/browse-movies/0/all/all/0/year/0/all?page=2" {
			close(ch)
			break
		}
		movieLinks, nextPage = y.getMovieLinkPerPage(nextPage)
	}
	return movies
}

func (y *Client) CollectMovieDetails(ch chan movie.Movie, done chan struct{}, wg *sync.WaitGroup, links []string) {
	var mu sync.Mutex
	for _, link := range links {
		wg.Add(1)
		go func(link string) {
			defer wg.Done()

			var movie movie.Movie
			movie.PageUrl = link
			movie.DownloadLinks = make(map[string]string)
			y.c.OnHTML(`div[id=mobile-movie-info]`, func(e *colly.HTMLElement) {
				e.ForEach(`h1`, func(i int, h *colly.HTMLElement) {
					// name
					movie.Name = h.Text
				})

				e.ForEach(`h2`, func(i int, h *colly.HTMLElement) {
					// year
					if i == 0 {
						movie.Year = strings.Split(h.Text, " ")[0]
					} else {
						movie.Categories = strings.Split(h.Text, "/")
					}
				})
			})
			y.c.OnHTML(`img.img-responsive`, func(img *colly.HTMLElement) {
				movie.ImageUrl = img.Attr("src")
			})
			y.c.OnHTML(`div[id=synopsis]`, func(d *colly.HTMLElement) {
				d.ForEach(`p.hidden-lg`, func(i int, p *colly.HTMLElement) {
					movie.Description = p.Text
				})
			})
			y.c.OnHTML(`div.bottom-info`, func(d *colly.HTMLElement) {
				d.ForEach(`p.hidden-md`, func(i int, p *colly.HTMLElement) {
					p.ForEach(`a`, func(i int, a *colly.HTMLElement) {
						mu.Lock()
						movie.DownloadLinks[a.Text] = a.Attr("href")
						mu.Unlock()
					})
				})
			})
			y.c.Visit(link)
			ch <- movie
		}(link)

	}
	wg.Wait()
	// done so we know to went through all the links in the page.
	done <- struct{}{}

}

// getMovieLinkPerPage gets the movies data from each page.
func (y *Client) getMovieLinkPerPage(url string) ([]string, string) {
	var movieLinks []string
	var nextPageUrl string

	// on every div with this class
	y.c.OnHTML(`a.browse-movie-link`, func(a *colly.HTMLElement) {
		movieLinks = append(movieLinks, a.Attr("href"))
	})
	// get next page link
	y.c.OnHTML(`a`, func(a *colly.HTMLElement) {
		if a.Text == "Next »" {
			nextPageUrl = fmt.Sprintf("https://%s%s", Domain, a.Attr("href"))
		}
	})
	y.c.Visit(url)
	return movieLinks, nextPageUrl
}
