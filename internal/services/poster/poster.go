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

	//	if err := p.SelectAndSendArticle(ctx); err != nil {
	//		return err
	//	}

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
	ext := ""
	// Проходим по всем файлам в папке
	err := filepath.Walk(imageDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Проверим, файл ли это (а не папка)
		if !info.IsDir() {
			// Получим расширение и проверим, изображение ли это
			ext = strings.ToLower(filepath.Ext(info.Name()))
			if ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif" {
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
	/*
		imageName := "pusik.jpg" // имя файла картинки
		imagePath := filepath.Join("images", imageName)

		// Открываем файл
		file, err := os.Open(imagePath)
		if err != nil {
			log.Println("Ошибка открытия картинки:", err)
		}
		defer file.Close()

		// Создаем объект фото
		photo := tgbotapi.NewPhoto(p.channelID, tgbotapi.FileReader{
			Name:   imageName,
			Reader: file,
			//	Size:   -1, // можно оставить -1, если не знаем размер
		})
		photo.Caption = "Вот твоя пикча, Пусичек! 💖"

		// Отправляем
		if _, err := p.bot.Send(imagePaths); err != nil {
			log.Println("Ошибка отправки фото:", err)
		}
	*/

	// Попробуем взять случайный и удалить
	img, ok := p.popRandom(&imagePaths)
	if ok {
		fmt.Println("Выбранная картинка:", img)
		fmt.Println("Оставшиеся картинки:", imagePaths)

		// Открываем файл
		file, err := os.Open(img)
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

		imgBase64, _ := helpers.EncodeImageToBase64(data, ext)

		// Создаем объект фото
		photo := tgbotapi.NewPhoto(p.channelID, tgbotapi.FileReader{
			Name:   img,
			Reader: file,
			//	Size:   -1, // можно оставить -1, если не знаем размер
		})
		photo.Caption, _ = p.openai.SetCaption("картинка", imgBase64)

		// Отправляем
		if _, err := p.bot.Send(photo); err != nil {
			log.Println("Ошибка отправки фото:", err)
		}

		err = os.Remove(img)
		if err != nil {
			fmt.Println("Ошибка при удалении файла:", err)
		} else {
			fmt.Println("Файл удалён успешно 🧼✨")
		}

		//	if err := p.sendPost(file); err != nil {
		//		return err
		//	}

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
