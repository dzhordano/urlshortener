package queries

import "github.com/dzhordano/urlshortener/internal/pkg/errs"

type RedirectQuery struct {
	ShortURL string
}

func NewRedirectQuery(shortURL string) (RedirectQuery, error) {
	if shortURL == "" {
		return RedirectQuery{}, errs.NewValueIsInvalidError("shortURL")
	}

	return RedirectQuery{
		ShortURL: shortURL,
	}, nil
}
