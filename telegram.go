package telegram // import "github.com/robertgzr/joe-telegram-adapter"

import (
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func parseChatID(channel string) (int64, error) {
	chatID, err := strconv.ParseInt(channel, 10, 64)
	if err != nil {
		return -1, fmt.Errorf("failed to parse chat id: %v", err)
	}
	return chatID, nil
}

func formatChatID(chatID int64) string {
	return strconv.FormatInt(chatID, 10)
}

func (a *TelegramAdapter) SendPhoto(channel string, photo interface{}, caption string) error {
	chatID, err := parseChatID(channel)
	if err != nil {
		return err
	}
	var cfg tgbotapi.PhotoConfig
	fileID, ok := photo.(string)
	if ok {
		cfg = tgbotapi.NewPhotoShare(chatID, fileID)
	} else {
		cfg = tgbotapi.NewPhotoUpload(chatID, photo)
	}
	if caption != "" {
		cfg.Caption = caption
	}
	_, err = a.BotAPI.Send(cfg)
	return err
}

func (a *TelegramAdapter) SendGIF(channel string, gif interface{}, caption string) error {
	chatID, err := parseChatID(channel)
	if err != nil {
		return err
	}
	var cfg tgbotapi.AnimationConfig
	fileID, ok := gif.(string)
	if ok {
		cfg = tgbotapi.NewAnimationShare(chatID, fileID)
	} else {
		cfg = tgbotapi.NewAnimationUpload(chatID, gif)
	}
	if caption != "" {
		cfg.Caption = caption
	}
	_, err = a.BotAPI.Send(cfg)
	return err
}

func (a *TelegramAdapter) SendSticker(channel string, sticker interface{}) error {
	chatID, err := parseChatID(channel)
	if err != nil {
		return err
	}
	var cfg tgbotapi.StickerConfig
	fileID, ok := sticker.(string)
	if ok {
		cfg = tgbotapi.NewStickerShare(chatID, fileID)
	} else {
		cfg = tgbotapi.NewStickerUpload(chatID, sticker)
	}
	_, err = a.BotAPI.Send(cfg)
	return err
}

type Callback = func(channel string) error
type Button struct {
	id    string
	label string
	cb    Callback
}

func (a *TelegramAdapter) NewButton(label string, cb Callback) Button {
	return Button{
		id:    "_button_" + strings.ToLower(label),
		label: label,
		cb:    cb,
	}
}

func (a *TelegramAdapter) SendButtons(channel, text string, buttons ...Button) error {
	chatID, err := parseChatID(channel)
	if err != nil {
		return err
	}

	var buttonRow []tgbotapi.InlineKeyboardButton
	for _, button := range buttons {
		a.callbacks[button.id] = button.cb
		buttonRow = append(buttonRow,
			tgbotapi.NewInlineKeyboardButtonData(button.label, button.id))
	}

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(buttonRow...))
	_, err = a.BotAPI.Send(msg)
	return err
}
