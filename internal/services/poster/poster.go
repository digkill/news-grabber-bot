package poster

import (
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/digkill/news-grabber-bot/internal/helpers"
	"github.com/digkill/news-grabber-bot/internal/summary"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	captionPrompt = "Generate a short Russian caption for this media (max 15 words). Avoid hashtags and links; optionally add one fitting emoji."
	postSignature = "Subscribe: https://t.me/noname_mem\nAI helper: https://t.me/AIVideoBestBot"
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
		log.Println("No images found in", p.imageDir)
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
		log.Printf("Error deleting file %s: %v", imgPath, err)
	} else {
		log.Printf("File deleted: %s", imgPath)
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
	if _, err = file.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("failed to rewind file: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(file.Name()))

	if ext == ".mp4" || ext == ".mov" {
		return p.sendVideo(file, data)
	}
	return p.sendPhoto(file, data, ext)
}

func (p *Poster) sendVideo(file *os.File, data []byte) error {
	outputPath := filepath.Join(p.imageDir, "frame.jpg")

	cmd := exec.Command("ffmpeg", "-i", file.Name(), "-frames:v", "1", outputPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg error: %w", err)
	}
	defer os.Remove(outputPath)

	previewData, err := os.ReadFile(outputPath)
	if err != nil {
		return fmt.Errorf("failed to read preview: %w", err)
	}

	imgBase64, err := helpers.EncodeImageToBase64(previewData, ".jpg")
	if err != nil {
		return fmt.Errorf("failed to encode preview: %w", err)
	}
	caption, err := p.openai.SetCaption(captionPrompt, imgBase64)
	if err != nil {
		log.Printf("failed to set caption for video: %v", err)
		caption = ""
	}

	videoMsg := tgbotapi.NewVideo(p.channelID, tgbotapi.FileReader{Name: file.Name(), Reader: file})
	videoMsg.Caption = strings.TrimSpace(caption + "\n\n" + postSignature)

	_, err = p.bot.Send(videoMsg)
	return err
}

func (p *Poster) sendPhoto(file *os.File, data []byte, ext string) error {
	imgBase64, err := helpers.EncodeImageToBase64(data, ext)
	if err != nil {
		return fmt.Errorf("failed to encode image: %w", err)
	}
	caption, err := p.openai.SetCaption(captionPrompt, imgBase64)
	if err != nil {
		log.Printf("failed to set caption for photo: %v", err)
		caption = ""
	}

	photoMsg := tgbotapi.NewPhoto(p.channelID, tgbotapi.FileReader{Name: file.Name(), Reader: file})
	photoMsg.Caption = strings.TrimSpace(caption + "\n\n" + postSignature)

	_, err = p.bot.Send(photoMsg)
	return err
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
