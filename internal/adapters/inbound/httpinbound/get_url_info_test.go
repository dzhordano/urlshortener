//nolint:nolintlint,exhaustruct,testpackage
package httpinbound

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dzhordano/urlshortener/internal/core/application/usecases/queries"
	queries_mocks "github.com/dzhordano/urlshortener/mocks/core/application_mocks/queries"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestServer_GetShortenedURLInfo(t *testing.T) {
	tt := []struct {
		name string
		// This is not good, however, i can afford that
		isAuthorized bool
		reqShortURL  string
		expectedCode int
		expectErr    bool
		mockBehavior func(m *queries_mocks.GetURLInfoQueryHandlerMock, q queries.GetURLInfoQuery)
	}{
		{
			name:         "success",
			isAuthorized: true,
			reqShortURL:  "RAND000",
			expectedCode: http.StatusOK,
			expectErr:    false,
			mockBehavior: func(m *queries_mocks.GetURLInfoQueryHandlerMock, q queries.GetURLInfoQuery) {
				m.On("Handle", mock.Anything, q).
					Return(queries.GetURLInfoResponse{
						OriginalURL:   "",
						ShortURL:      "",
						Clicks:        0,
						CreatedAtUTC:  time.Time{},
						ValidUntilUTC: time.Time{},
					}, nil).
					Once()
			},
		},
		{
			name:         "bad request",
			isAuthorized: true,
			reqShortURL:  "",
			expectedCode: http.StatusBadRequest,
			expectErr:    true,
			mockBehavior: func(*queries_mocks.GetURLInfoQueryHandlerMock, queries.GetURLInfoQuery) {},
		},
		{
			name:         "internal",
			isAuthorized: true,
			reqShortURL:  "EXISTS0",
			expectedCode: http.StatusInternalServerError,
			expectErr:    true,
			mockBehavior: func(m *queries_mocks.GetURLInfoQueryHandlerMock, q queries.GetURLInfoQuery) {
				m.On("Handle", mock.Anything, q).
					Return(queries.GetURLInfoResponse{}, assert.AnError).
					Once()
			},
		},
		{
			name:         "unauthorized",
			isAuthorized: false,
			reqShortURL:  "RAND000",
			expectedCode: http.StatusUnauthorized,
			expectErr:    true,
			mockBehavior: func(*queries_mocks.GetURLInfoQueryHandlerMock, queries.GetURLInfoQuery) {},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(
				http.MethodGet,
				fmt.Sprintf("/api/v1/%s/info", tc.reqShortURL),
				nil,
			)
			rec := httptest.NewRecorder()
			ctx := e.NewContext(req, rec)
			if tc.isAuthorized {
				ctx.Request().Header.Set("X-Api-Key", "admin")
			}

			m := queries_mocks.NewGetURLInfoQueryHandlerMock(t)
			q := queries.GetURLInfoQuery{ShortURL: tc.reqShortURL}
			tc.mockBehavior(m, q)

			s := &Server{
				shortenURLCommandHandler: nil,
				redirectQueryHandler:     nil,
				getURLInfoQueryHandler:   m,
			}

			err := s.GetShortenedURLInfo(ctx, tc.reqShortURL)

			if err != nil {
				// Since echo.NewHTTPError() is used to return errors, cast error to echo.HTTPError
				var httpErr *echo.HTTPError
				if errors.As(err, &httpErr) {
					assert.Equal(t, tc.expectedCode, httpErr.Code)
				} else if tc.expectErr {
					assert.Error(t, err)
				}
			} else {
				assert.Equal(t, tc.expectedCode, rec.Code)
			}
		})
	}
}
