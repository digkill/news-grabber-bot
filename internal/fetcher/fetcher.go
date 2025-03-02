package fetcher

import (
	"context"
	"github.com/digkill/news-grabber-bot/internal/models"
	src "github.com/digkill/news-grabber-bot/internal/source"
	"go.tomakado.io/containers/set"
	"log"
	"strings"
	"sync"
	"time"
)

//go:generate moq --out=mocks/mock_article_storage.go --pkg=mocks . ArticleStorage
type ArticleStorage interface {
	Store(ctx context.Context, article models.Article) error
}

//go:generate moq --out=mocks/mock_sources_provider.go --pkg=mocks . SourcesProvider
type SourcesProvider interface {
	Sources(ctx context.Context) ([]models.Source, error)
}

type Source interface {
	ID() int64
	Name() string
	Fetch(ctx context.Context) ([]models.Item, error)
}

type Fetcher struct {
	articles ArticleStorage
	sources  SourcesProvider

	fetchInterval  time.Duration
	filterKeywords []string
}

func NewFetcher(
	articles ArticleStorage,
	sources SourcesProvider,
	fetchInterval time.Duration,
	filterKeywords []string,
) *Fetcher {
	return &Fetcher{
		articles:       articles,
		sources:        sources,
		fetchInterval:  fetchInterval,
		filterKeywords: filterKeywords,
	}
}

func (f *Fetcher) Fetch(ctx context.Context) error {
	sources, err := f.sources.Sources(ctx)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup

	for _, source := range sources {
		wg.Add(1)

		go func(source Source) {
			defer wg.Done()

			items, err := source.Fetch(ctx)
			if err != nil {
				log.Printf("[ERROR] failed to fetch items from source %q: %v", source.Name(), err)
				return
			}

			if err := f.processItems(ctx, source, items); err != nil {
				log.Printf("[ERROR] failed to process items from source %q: %v", source.Name(), err)
				return
			}

		}(src.NewRSSSourceFromModel(source))
	}

	wg.Wait()

	return nil
}

func (f *Fetcher) Start(ctx context.Context) error {
	ticker := time.NewTicker(f.fetchInterval)
	defer ticker.Stop()

	if err := f.Fetch(ctx); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := f.Fetch(ctx); err != nil {
				return err
			}
		}
	}
}

func (f *Fetcher) processItems(ctx context.Context, source Source, items []models.Item) error {
	for _, item := range items {
		item.Date = item.Date.UTC()

		if f.itemShouldBeSkipped(item) {
			log.Printf("[INFO] item %q (%s) from source %q should be skipped", item.Title, item.Link, source.Name())
			continue
		}

		if err := f.articles.Store(ctx, models.Article{
			SourceID:    source.ID(),
			Title:       item.Title,
			Link:        item.Link,
			Summary:     item.Summary,
			PublishedAt: item.Date,
		}); err != nil {
			return err
		}
	}

	return nil
}

func (f *Fetcher) itemShouldBeSkipped(item models.Item) bool {
	categoriesSet := set.New(item.Categories...)

	for _, keyword := range f.filterKeywords {
		if categoriesSet.Contains(keyword) || strings.Contains(strings.ToLower(item.Title), keyword) {
			return true
		}
	}

	return false
}
