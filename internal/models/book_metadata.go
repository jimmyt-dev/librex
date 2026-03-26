package models

type BookMetadata struct {
	BookID        string   `json:"bookId"`
	Title         string   `json:"title"`
	Subtitle      *string  `json:"subtitle"`
	Description   *string  `json:"description"`
	Publisher     *string  `json:"publisher"`
	PublishedDate *string  `json:"publishedDate"`
	ISBN13        *string  `json:"isbn13"`
	ISBN10        *string  `json:"isbn10"`
	Language      *string  `json:"language"`
	PageCount     *int     `json:"pageCount"`
	SeriesName    *string  `json:"seriesName"`
	SeriesNumber  *float64 `json:"seriesNumber"`
	SeriesTotal   *int     `json:"seriesTotal"`
	Rating        *int     `json:"rating"`
	CoverPath     *string  `json:"coverPath"`
	CoverMime     *string  `json:"coverMime"`
}
