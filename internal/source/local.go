package source

import (
	"context"
	"github.com/digkill/news-grabber-bot/internal/models"
	"log"
	"os"
	"path/filepath"
)

type LocalSource struct {
	File       *os.File
	SourceID   int64
	SourceName string
}

func NewLocalSourceFromModel(m models.File) LocalSource {
	return LocalSource{
		File:       m.File,
		SourceID:   m.SourceID,
		SourceName: m.SourceName,
	}
}

func (s LocalSource) Fetch(ctx context.Context) (*os.File, error) {
	feed, err := s.loadFeed(ctx, s.URL)
	if err != nil {
		return nil, err
	}

	return feed, nil

	//	return lo.Map(feed.Items, func(item *rss.Item, _ int) models.Item {
	//		return models.Item{
	//			Title:      item.Title,
	//			Link:       item.Link,
	//			Categories: item.Categories,
	//			Date:       item.Date,
	//			SourceName: s.SourceName,
	//			Summary:    strings.TrimSpace(item.Summary),
	//		}
	//	}), nil
}

func (s LocalSource) ID() int64 {
	return s.SourceID
}

func (s LocalSource) Name() string {
	return s.SourceName
}

func (s LocalSource) loadFeed(ctx context.Context, url string) (*os.File, error) {
	var feedCh = make(chan *os.File)
	var errCh = make(chan error)

	imageName := "image.gif" // имя файла картинки
	imagePath := filepath.Join("./internal/storage/images", imageName)

	// Открываем файл
	file, err := os.Open(imagePath)
	if err != nil {
		log.Println("Ошибка открытия картинки:", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Println("Ошибка закрытие файла:", err)
		}
	}(file)

	go func() {

		feed := file
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
