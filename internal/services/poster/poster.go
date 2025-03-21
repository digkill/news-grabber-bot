package poster

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"os"
	"time"
)

type PosterInterface interface {
	Posting(ctx context.Context) error
}

type Poster struct {
	files         []*os.File
	fetchInterval time.Duration
	bot           *tgbotapi.BotAPI
}

func (f *Poster) Posting(ctx context.Context) error {

	return nil
}

func NewPoster(
	files []*os.File,
	fetchInterval time.Duration,

) *Poster {
	return &Poster{
		files:         files,
		fetchInterval: fetchInterval,
	}
}
