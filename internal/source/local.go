package source

import (
	"context"

	"github.com/digkill/news-grabber-bot/internal/models"
)

type LocalSource struct {
	URL        string
	SourceID   int64
	SourceName string
}

func NewLocalSourceFromModel(m models.Source) LocalSource {
	return LocalSource{
		URL:        m.FeedURL,
		SourceID:   m.ID,
		SourceName: m.Name,
	}
}

func (s LocalSource) Fetch(_ context.Context) ([]models.Item, error) {
	return nil, nil
}

func (s LocalSource) ID() int64 {
	return s.SourceID
}

func (s LocalSource) Name() string {
	return s.SourceName
}
