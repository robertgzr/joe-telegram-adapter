package telegram // import "github.com/robertgzr/joe-telegram-adapter"

import (
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type ReceiveCommandEvent struct {
	Arg0 string
	Args []string
	From *tgbotapi.User
	Chat *tgbotapi.Chat
	Data *tgbotapi.Message
}

func (e ReceiveCommandEvent) Channel() string {
	return strconv.FormatInt(e.Chat.ID, 10)
}
