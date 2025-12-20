//nolint:nolintlint,exhaustruct,testpackage
package httpinbound

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dzhordano/urlshortener/internal/core/application/usecases/queries"
	"github.com/dzhordano/urlshortener/internal/pkg/errs"
	queries_mocks "github.com/dzhordano/urlshortener/mocks/core/application_mocks/queries"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestServer_Redirect(t *testing.T) {
	tt := []struct {
		name         string
		reqShortURL  string
		expectedCode int
		expectErr    bool
		mockBehavior func(m *queries_mocks.RedirectQueryHandlerMock, q queries.RedirectQuery)
	}{
		{
			name:         "success",
			reqShortURL:  "RAND000",
			expectedCode: http.StatusMovedPermanently,
			expectErr:    false,
			mockBehavior: func(m *queries_mocks.RedirectQueryHandlerMock, q queries.RedirectQuery) {
				m.On("Handle", mock.Anything, q).
					Return(queries.RedirectResponse{}, nil).
					Once()
			},
		},
		{
			name:         "bad request",
			reqShortURL:  "",
			expectedCode: http.StatusBadRequest,
			expectErr:    true,
			mockBehavior: func(*queries_mocks.RedirectQueryHandlerMock, queries.RedirectQuery) {},
		},
		{
			name:         "internal",
			reqShortURL:  "EXISTS0",
			expectedCode: http.StatusInternalServerError,
			expectErr:    true,
			mockBehavior: func(m *queries_mocks.RedirectQueryHandlerMock, q queries.RedirectQuery) {
				m.On("Handle", mock.Anything, q).
					Return(queries.RedirectResponse{}, assert.AnError).
					Once()
			},
		},
		{
			name:         "not found",
			reqShortURL:  "NOTFND0",
			expectedCode: http.StatusNotFound,
			expectErr:    false,
			mockBehavior: func(m *queries_mocks.RedirectQueryHandlerMock, q queries.RedirectQuery) {
				m.On("Handle", mock.Anything, q).
					Return(queries.RedirectResponse{}, errs.ErrObjectNotFound).
					Once()
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(
				http.MethodGet,
				fmt.Sprintf("/api/v1/%s", tc.reqShortURL),
				nil,
			)
			rec := httptest.NewRecorder()
			ctx := e.NewContext(req, rec)

			m := queries_mocks.NewRedirectQueryHandlerMock(t)
			q := queries.RedirectQuery{ShortURL: tc.reqShortURL}
			tc.mockBehavior(m, q)

			s := &Server{
				redirectQueryHandler: m,
			}

			err := s.Redirect(ctx, tc.reqShortURL)

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
