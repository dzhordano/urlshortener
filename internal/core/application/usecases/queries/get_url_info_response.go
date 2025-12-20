package queries

import "time"

type GetURLInfoResponse struct {
	OriginalURL  string
	ShortURL     string
	Clicks       int
	CreatedAtUTC time.Time
}
