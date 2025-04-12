package poster

import (
	"bytes"
	"context"
	"fmt"
	"github.com/digkill/news-grabber-bot/internal/helpers"
	"github.com/digkill/news-grabber-bot/internal/summary"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"io"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Poster struct {
	imageDir     string
	postInterval time.Duration
	bot          *tgbotapi.BotAPI
	channelID    int64
	openai       *summary.OpenAI
}

func NewPoster(imageDir string, postInterval time.Duration, bot *tgbotapi.BotAPI, channelID int64, openai *summary.OpenAI) *Poster {
	return &Poster{
		imageDir:     imageDir,
		postInterval: postInterval,
		bot:          bot,
		channelID:    channelID,
		openai:       openai,
	}
}

func (p *Poster) Start(ctx context.Context) error {
	ticker := time.NewTicker(p.postInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := p.Posting(ctx); err != nil {
				log.Println("Error during posting:", err)
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (p *Poster) Posting(ctx context.Context) error {
	images, err := p.getImagePaths()
	if err != nil {
		return fmt.Errorf("failed to scan images: %w", err)
	}

	if len(images) == 0 {
		log.Println("No images found 😿")
		return nil
	}

	imgPath := p.pickRandomImage(images)
	log.Println("Selected image:", imgPath)

	if err := p.isValidImage(imgPath); err != nil {
		return err
	}

	if err = p.processAndSendImage(imgPath); err != nil {
		return fmt.Errorf("failed to process image %s: %w", imgPath, err)
	}

	if err = os.Remove(imgPath); err != nil {
		log.Println("Error deleting file:", err)
	} else {
		log.Println("File deleted successfully 🧼✨")
	}

	return nil
}

func (p *Poster) getImagePaths() ([]string, error) {
	var images []string

	err := filepath.Walk(p.imageDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		if ext := strings.ToLower(filepath.Ext(info.Name())); helpers.IsImageOrVideo(ext) {
			images = append(images, path)
		}
		return nil
	})

	return images, err
}

func (p *Poster) pickRandomImage(images []string) string {
	rand.Seed(time.Now().UnixNano())
	return images[rand.Intn(len(images))]
}

func (p *Poster) processAndSendImage(imgPath string) error {
	file, err := os.Open(imgPath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(file.Name()))

	if len(data) == 0 {
		return fmt.Errorf("file is empty: %s", imgPath)
	}

	if ext == ".mp4" {
		return p.sendVideo(file, data)
	}
	return p.sendPhoto(file, data, ext)
}

func (p *Poster) sendVideo(file *os.File, data []byte) error {
	reader := bytes.NewReader(data)

	videoFile := tgbotapi.FileReader{
		Name:   filepath.Base(file.Name()),
		Reader: reader,
	}

	msg := tgbotapi.NewVideo(p.channelID, videoFile)
	msg.Caption = "Лови видосик 🎥"
	_, err := p.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("failed to send video: %w", err)
	}
	return nil
}

func (p *Poster) sendPhoto(file *os.File, data []byte, ext string) error {
	reader := bytes.NewReader(data)

	photoFile := tgbotapi.FileReader{
		Name:   filepath.Base(file.Name()),
		Reader: reader,
	}

	msg := tgbotapi.NewPhoto(p.channelID, photoFile)
	msg.Caption = "Вот тебе картинка 😽"
	_, err := p.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("failed to send photo: %w", err)
	}
	return nil
}

func (p *Poster) isValidImage(path string) error {
	stat, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat image: %w", err)
	}
	if stat.Size() == 0 {
		return fmt.Errorf("image is empty: %s", path)
	}
	return nil
}
