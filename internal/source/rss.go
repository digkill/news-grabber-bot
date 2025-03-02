package source

import (
	"context"
	"github.com/digkill/news-grabber-bot/internal/models"
	"strings"

	"github.com/SlyMarbo/rss"
	"github.com/samber/lo"
)

type RSSSource struct {
	URL        string
	SourceID   int64
	SourceName string
}

func NewRSSSourceFromModel(m models.Source) RSSSource {
	return RSSSource{
		URL:        m.FeedURL,
		SourceID:   m.ID,
		SourceName: m.Name,
	}
}

func (s RSSSource) Fetch(ctx context.Context) ([]models.Item, error) {
	feed, err := s.loadFeed(ctx, s.URL)
	if err != nil {
		return nil, err
	}

	return lo.Map(feed.Items, func(item *rss.Item, _ int) models.Item {
		return models.Item{
			Title:      item.Title,
			Link:       item.Link,
			Categories: item.Categories,
			Date:       item.Date,
			SourceName: s.SourceName,
			Summary:    strings.TrimSpace(item.Summary),
		}
	}), nil
}

func (s RSSSource) ID() int64 {
	return s.SourceID
}

func (s RSSSource) Name() string {
	return s.SourceName
}

func (s RSSSource) loadFeed(ctx context.Context, url string) (*rss.Feed, error) {
	var feedCh = make(chan *rss.Feed)
	var errCh = make(chan error)

	go func() {
		feed, err := rss.Fetch(url)
		if err != nil {
			errCh <- err
			return
		}
		feedCh <- feed
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case err := <-errCh:
		return nil, err
	case feed := <-feedCh:
		return feed, nil
	}
}
