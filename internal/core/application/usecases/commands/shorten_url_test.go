//nolint:nolintlint,exhaustruct,testpackage
package commands

import (
	"context"
	"testing"

	"github.com/dzhordano/urlshortener/internal/pkg/errs"
	"github.com/dzhordano/urlshortener/internal/pkg/logger"
	"github.com/dzhordano/urlshortener/mocks/core/ports_mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestShortenURLCommandHandler_SuccessSaved(t *testing.T) {
	ctx := context.Background()
	cmd := ShortenURLCommand{
		OriginalURL: "https://example.com",
	}

	rm := ports_mocks.NewURLRepositoryMock(t)
	cm := ports_mocks.NewURLCacheMock(t)
	l, err := logger.NewSlogLogger(true, "debug")
	assert.NoError(t, err)

	rm.On("Save", mock.Anything, mock.Anything).Return(nil).Once()
	cm.On("Set", mock.Anything, mock.Anything, mock.Anything).Return(nil).Times(1)

	ch, _ := NewShortenURLCommandHandler(l, cm, rm)
	resp, err := ch.Handle(ctx, cmd)

	require.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestShortenURLCommandHandler_InvalidCommand(t *testing.T) {
	ctx := context.Background()
	cmd := ShortenURLCommand{
		OriginalURL: "",
	}

	rm := ports_mocks.NewURLRepositoryMock(t)
	cm := ports_mocks.NewURLCacheMock(t)
	l, err := logger.NewSlogLogger(true, "debug")
	assert.NoError(t, err)

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
	cm := ports_mocks.NewURLCacheMock(t)
	l, err := logger.NewSlogLogger(true, "debug")
	assert.NoError(t, err)

	rm.On("Save", mock.Anything, mock.Anything).Return(assert.AnError).Once()

	ch, _ := NewShortenURLCommandHandler(l, cm, rm)
	resp, err := ch.Handle(ctx, cmd)

	require.ErrorIs(t, assert.AnError, err)
	assert.Empty(t, resp)
}
