package bot

import (
	"context"
	"fmt"
	"github.com/digkill/news-grabber-bot/internal/botkit"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type CmdStart interface {
}

func ViewCmdStart(prioritySetter CmdStart) botkit.ViewFunc {

	return func(ctx context.Context, bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Приветствую!")

		if _, err := bot.Send(msg); err != nil {
			fmt.Println(err)
			return err
		}

		return nil
	}
}
