package movie

type Movie struct {
	ID            int               `json:"id"`
	Categories    []string          `json:"catagories"`
	Name          string            `json:"name"`
	ImageUrl      string            `json:"image"`
	PageUrl       string            `json:"page_url"`
	Year          string            `json:"year"`
	Description   string            `json:"description"`
	DownloadLinks map[string]string `json:"download_links"`
}
