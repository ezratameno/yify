package movie

type Movie struct {
	ID            int               `json:"id"`
	Categories    []string          `json:"catagories"`
	Name          string            `json:"name"`
	ImageUrl      string            `json:"image_url"`
	PageUrl       string            `json:"page_url"`
	Year          string            `json:"year"`
	Description   string            `json:"description"`
	DownloadLinks map[string]string `json:"download_links"`
}

// NewUser contains information needed to create a new Movie.
type NewMovie struct {
	ID            int    `json:"id"`
	Categories    string `json:"catagories"`
	Name          string `json:"name"`
	ImageUrl      string `json:"image"`
	PageUrl       string `json:"page_url"`
	Year          string `json:"year"`
	Description   string `json:"description"`
	DownloadLinks string `json:"download_links"`
}
