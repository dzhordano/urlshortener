//nolint:nolintlint,exhaustruct,testpackage
package httpinbound

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dzhordano/urlshortener/internal/core/application/usecases/commands"
	"github.com/dzhordano/urlshortener/internal/gen/servers"
	commands_mocks "github.com/dzhordano/urlshortener/mocks/core/application_mocks/commands"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestServer_ShortenURL(t *testing.T) {
	tt := []struct {
		name           string
		reqOriginalURL string
		expectedCode   int
		expectErr      bool
		mockBehavior   func(m *commands_mocks.ShortenURLCommandHandlerMock, c commands.ShortenURLCommand)
	}{
		{
			name:           "success",
			reqOriginalURL: "https://google.com",
			expectedCode:   http.StatusOK,
			expectErr:      false,
			mockBehavior: func(m *commands_mocks.ShortenURLCommandHandlerMock, c commands.ShortenURLCommand) {
				m.On("Handle", mock.Anything, c).
					Return("SHORT00", nil).
					Once()
			},
		},
		{
			name:           "bad request",
			reqOriginalURL: "",
			expectedCode:   http.StatusBadRequest,
			expectErr:      true,
			mockBehavior:   func(*commands_mocks.ShortenURLCommandHandlerMock, commands.ShortenURLCommand) {},
		},
		{
			name:           "internal",
			reqOriginalURL: "https://google.com",
			expectedCode:   http.StatusInternalServerError,
			expectErr:      true,
			mockBehavior: func(m *commands_mocks.ShortenURLCommandHandlerMock, c commands.ShortenURLCommand) {
				m.On("Handle", mock.Anything, c).
					Return("", assert.AnError).
					Once()
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()
			rs := servers.ShortenURLJSONBody{
				Url: tc.reqOriginalURL,
			}
			body, _ := json.Marshal(rs)
			req := httptest.NewRequest(http.MethodGet, "/", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			ctx := e.NewContext(req, rec)

			m := commands_mocks.NewShortenURLCommandHandlerMock(t)
			c := commands.ShortenURLCommand{OriginalURL: tc.reqOriginalURL}
			tc.mockBehavior(m, c)

			s := &Server{
				shortenURLCommandHandler: m,
				redirectQueryHandler:     nil,
				getURLInfoQueryHandler:   nil,
			}

			err := s.ShortenURL(ctx)

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
