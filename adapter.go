package telegram // import "github.com/robertgzr/joe-telegram-adapter"

import (
	"context"
	"strconv"
	"strings"

	"github.com/go-joe/joe"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type TelegramAdapter struct {
	context context.Context
	logger  *zap.Logger
	name    string
	userID  int

	tg      *tgbotapi.BotAPI
	updates tgbotapi.UpdatesChannel
}

type Config struct {
	Token            string
	UpdateTimeoutSec int
	Logger           *zap.Logger
}

func Adapter(token string, opts ...Option) joe.Module {
	return joe.ModuleFunc(func(joeConf *joe.Config) error {
		conf := Config{Token: token}

		for _, opt := range opts {
			err := opt(&conf)
			if err != nil {
				return err
			}
		}

		if conf.Logger == nil {
			conf.Logger = joeConf.Logger("telegram")
		}

		a, err := NewAdapter(joeConf.Context, conf)
		if err != nil {
			return err
		}

		joeConf.SetAdapter(a)
		return nil
	})
}

func NewAdapter(ctx context.Context, conf Config) (*TelegramAdapter, error) {
	tg, err := tgbotapi.NewBotAPI(conf.Token)
	if err != nil {
		return nil, errors.Wrap(err, "telegram failed to initialize")
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = conf.UpdateTimeoutSec
	updates, err := tg.GetUpdatesChan(u)
	if err != nil {
		return nil, errors.Wrap(err, "telegram failed to get updates")
	}

	return newAdapter(ctx, tg, updates, conf)
}

func newAdapter(ctx context.Context, tg *tgbotapi.BotAPI, updates tgbotapi.UpdatesChannel, conf Config) (*TelegramAdapter, error) {
	a := &TelegramAdapter{
		tg:      tg,
		updates: updates,
		context: ctx,
		logger:  conf.Logger,
	}

	if a.logger == nil {
		a.logger = zap.NewNop()
	}

	user, err := tg.GetMe()
	if err != nil {
		return nil, errors.Wrap(err, "telegram failed to get bot user")
	}

	a.userID = user.ID
	a.logger.Info("Connected to telegram API",
		zap.String("user", user.UserName),
		zap.Int("user_id", user.ID),
	)
	return a, nil
}

// RegisterAt implements the joe.Adapter interface by emitting the telegram API
// events to the given brain
func (a *TelegramAdapter) RegisterAt(brain *joe.Brain) {
	go a.handleTelegramEvents(brain)
}

func (a *TelegramAdapter) handleTelegramEvents(brain *joe.Brain) {
	for update := range a.updates {
		select {
		case <-a.context.Done():
			a.logger.Debug("Cancelling event loop")
			return
		default:
		}

		if update.Message == nil {
			continue
		}

		m := update.Message
		text := strings.TrimSpace(m.Text)

		a.logger.Debug("Received message",
			zap.Int("update_id", update.UpdateID),
			zap.Int("message_id", m.MessageID),
		)

		if m.IsCommand() {
			arg0, argStr := m.Command(), m.CommandArguments()
			args := strings.Split(argStr, " ")

			a.logger.Debug("Received command",
				zap.String("command", arg0),
				zap.String("arguments", argStr),
			)

			brain.Emit(ReceiveCommandEvent{
				Arg0: arg0,
				Args: args,
				From: m.From,
				Chat: m.Chat,
				Data: m,
			})
		}

		brain.Emit(joe.ReceiveMessageEvent{
			Text:     text,
			Channel:  strconv.FormatInt(m.Chat.ID, 10),
			AuthorID: strconv.Itoa(m.From.ID),
			Data:     m,
		})
	}
}

func (a *TelegramAdapter) Send(txt, chatIDString string) error {
	chatID, err := strconv.ParseInt(chatIDString, 10, 64)
	if err != nil {
		return errors.Wrap(err, "failed to parse chat id")
	}

	a.logger.Info("Sending message to chat",
		zap.Int64("chat_id", chatID),
	)

	_, err = a.tg.Send(tgbotapi.NewMessage(chatID, txt))
	return err
}

func (a *TelegramAdapter) Close() error {
	a.updates.Clear()
	return nil
}
