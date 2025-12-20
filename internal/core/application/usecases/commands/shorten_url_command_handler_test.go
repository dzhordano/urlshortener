//nolint:nolintlint,exhaustruct,testpackage
package commands

import (
	"context"
	"testing"
	"time"

	"github.com/dzhordano/urlshortener/internal/core/domain/model"
	"github.com/dzhordano/urlshortener/internal/pkg/errs"
	"github.com/dzhordano/urlshortener/mocks/core/ports_mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestShortenURLCommandHandler_SuccessSaved(t *testing.T) {
	ctx := context.Background()
	cmd := ShortenURLCommand{
		OriginalURL: "https://example.com",
	}

	rm := ports_mocks.NewURLRepositoryMock(t)
	cm := ports_mocks.NewCacheMock(t)
	l := zaptest.NewLogger(t).Sugar()

	rm.On("Save", mock.Anything, mock.Anything).Return(nil).Once()
	cm.On("Set", mock.Anything, mock.Anything, mock.Anything).Return(nil).Times(1)

	ch, _ := NewShortenURLCommandHandler(l, cm, rm)
	resp, err := ch.Handle(ctx, cmd)

	require.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestShortenURLCommandHandler_SuccessAlreadyExists(t *testing.T) {
	ctx := context.Background()

	cmd := ShortenURLCommand{
		OriginalURL: "https://example.com",
	}

	rm := ports_mocks.NewURLRepositoryMock(t)
	cm := ports_mocks.NewCacheMock(t)
	l := zaptest.NewLogger(t).Sugar()

	rm.On("Save", mock.Anything, mock.Anything).Return(errs.ErrObjectAlreadyExists).Once()
	rm.On("GetByOriginalURL", mock.Anything, mock.Anything).Return(&model.ShortenedURL{
		OriginalURL:  "",
		ShortURL:     "SHORT01",
		Clicks:       0,
		CreatedAtUTC: time.Time{},
	}, nil).Once()
	cm.On("Set", mock.Anything, mock.Anything, mock.Anything).Return(nil).Times(1)

	ch, _ := NewShortenURLCommandHandler(l, cm, rm)
	resp, err := ch.Handle(ctx, cmd)

	require.NoError(t, err)
	assert.Equal(t, "SHORT01", resp)
}

func TestShortenURLCommandHandler_InvalidCommand(t *testing.T) {
	ctx := context.Background()
	cmd := ShortenURLCommand{
		OriginalURL: "",
	}

	rm := ports_mocks.NewURLRepositoryMock(t)
	cm := ports_mocks.NewCacheMock(t)
	l := zaptest.NewLogger(t).Sugar()

	ch, _ := NewShortenURLCommandHandler(l, cm, rm)
	resp, err := ch.Handle(ctx, cmd)

	require.ErrorIs(t, err, errs.ErrValueIsRequired)
	assert.Empty(t, resp)
}

func TestShortenURLCommandHandler_Internal(t *testing.T) {
	ctx := context.Background()
	cmd := ShortenURLCommand{
		OriginalURL: "https://example.com",
	}

	rm := ports_mocks.NewURLRepositoryMock(t)
	cm := ports_mocks.NewCacheMock(t)
	l := zaptest.NewLogger(t).Sugar()

	rm.On("Save", mock.Anything, mock.Anything).Return(assert.AnError).Once()

	ch, _ := NewShortenURLCommandHandler(l, cm, rm)
	resp, err := ch.Handle(ctx, cmd)

	require.ErrorIs(t, assert.AnError, err)
	assert.Empty(t, resp)
}
