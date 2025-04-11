package poster

import (
	"context"
	"fmt"
	"github.com/digkill/news-grabber-bot/internal/helpers"
	"github.com/digkill/news-grabber-bot/internal/summary"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"io"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Poster struct {
	pathImage    string
	postInterval time.Duration
	bot          *tgbotapi.BotAPI
	channelID    int64
	openai       *summary.OpenAI
}

// Start Запуск сервиса
func (p *Poster) Start(ctx context.Context) error {
	ticker := time.NewTicker(p.postInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := p.Posting(ctx); err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (p *Poster) Posting(ctx context.Context) error {

	imageDir := "./internal/storage/images"
	var imagePaths []string

	// Проходим по всем файлам в папке
	err := filepath.Walk(imageDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Проверим, файл ли это (а не папка)
		if !info.IsDir() {
			// Получим расширение и проверим, изображение ли это
			ext := strings.ToLower(filepath.Ext(info.Name()))
			if ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif" || ext == ".mp4" {
				imagePaths = append(imagePaths, path)
			}
		}
		return nil
	})

	if err != nil {
		fmt.Println("Ошибка сканирования папки:", err)
	}

	// Выведем список найденных изображений
	fmt.Println("Найденные изображения:")
	for _, img := range imagePaths {
		fmt.Println("-", img)
	}

	// Попробуем взять случайный и удалить
	img, ok := p.popRandom(&imagePaths)
	if ok {
		fmt.Println("Выбранная картинка:", img)
		fmt.Println("Оставшиеся картинки:", imagePaths)

		// Открываем файл
		file, _ := os.Open(img)
		if err != nil {
			log.Println("Ошибка открытия картинки:", err)
		}
		defer file.Close()

		p.openai.GetClient()

		// Преобразуем в []byte
		data, err := io.ReadAll(file)
		if err != nil {
			fmt.Println("Ошибка чтения файла:", err)
		}
		file.Seek(0, io.SeekStart)

		extF := strings.ToLower(filepath.Ext(file.Name()))
		if extF == ".mp4" {

			outputPath := "./internal/storage/images/frame.jpg"

			// Аргументы ffmpeg для взятия первого кадра
			cmd := exec.Command("ffmpeg", "-i", file.Name(), "-frames:v", "1", outputPath)

			if err = cmd.Run(); err != nil {
				fmt.Println("Ошибка:", err)
			}

			video := tgbotapi.NewVideo(p.channelID, tgbotapi.FileReader{
				Name:   file.Name(),
				Reader: file,
			})
			video.Caption, _ = p.openai.SetCaption("картинка мем", outputPath)

			// Отправляем
			if _, err = p.bot.Send(video); err != nil {
				log.Println("Ошибка отправки фото:", err)
			}
		} else {
			imgBase64, _ := helpers.EncodeImageToBase64(data, extF)

			// Создаем объект фото
			photo := tgbotapi.NewPhoto(p.channelID, tgbotapi.FileReader{
				Name:   file.Name(),
				Reader: file,
			})
			photo.Caption, _ = p.openai.SetCaption("картинка мем", imgBase64)

			// Отправляем
			if _, err = p.bot.Send(photo); err != nil {
				log.Println("Ошибка отправки фото:", err)
			}
		}

		//	godump.Dump(photo)

		err = os.Remove(img)
		if err != nil {
			fmt.Println("Ошибка при удалении файла:", err)
		} else {
			fmt.Println("Файл удалён успешно 🧼✨")
		}

	} else {
		fmt.Println("Слайс пустой 😿")
	}

	return nil
}

func (p *Poster) popRandom(images *[]string) (string, bool) {
	if len(*images) == 0 {
		return "", false
	}

	rand.Seed(time.Now().UnixNano())
	index := rand.Intn(len(*images))
	chosen := (*images)[index]

	// Удалим элемент из слайса
	*images = append((*images)[:index], (*images)[index+1:]...)

	return chosen, true
}

func (p *Poster) sendPost(images *os.File) error {
	//const msgFormat = "*%s*%s\n\n%s"

	msg := tgbotapi.NewMessage(p.channelID, "Тестовый пост")
	msg.ParseMode = "MarkdownV2"

	_, err := p.bot.Send(msg)
	if err != nil {
		return err
	}

	return nil
}

func NewPoster(
	pathImage string,
	postInterval time.Duration,
	bot *tgbotapi.BotAPI,
	channelID int64,
	openai *summary.OpenAI,
) *Poster {
	return &Poster{
		pathImage:    pathImage,
		postInterval: postInterval,
		bot:          bot,
		channelID:    channelID,
		openai:       openai,
	}
}
