// Package telegram implements a joe-bot adapter for the Telegram BotAPI
//
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

// TelegramAdapter is a joe-bot adapter for the Telegram BotAPI
type TelegramAdapter struct {
	context context.Context
	logger  *zap.Logger
	name    string
	userID  int
	conf    Config

	BotAPI  *tgbotapi.BotAPI
	updates tgbotapi.UpdatesChannel

	callbacks map[string]Callback
}

// Config holds the options pased to the Telegram adapter
type Config struct {
	// Token is the token issued by botfather
	// https://core.telegram.org/bots/api#authorizing-your-bot
	Token string
	// UpdateTimeoutSec configures the UpdateCofnig.Timout option of the
	// telegram-bot-api client
	UpdateTimeoutSec int
	// UpdateResumeFrom is the last Update ID to resume from
	// https://core.telegram.org/bots/api#update
	UpdateResumeFrom int
	// ParseMode sets the formatting style for messages the bot sends
	// https://core.telegram.org/bots/api#formatting-options
	ParseMode string

	Logger *zap.Logger
}

// Adapter returns a joe.Module of the TelegramAdapter
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

// NewAdapter creates a new instance of the TelegramAdapter
func NewAdapter(ctx context.Context, conf Config) (*TelegramAdapter, error) {
	tg, err := tgbotapi.NewBotAPI(conf.Token)
	if err != nil {
		return nil, errors.Wrap(err, "telegram failed to initialize")
	}

	u := tgbotapi.NewUpdate(conf.UpdateResumeFrom)
	u.Timeout = conf.UpdateTimeoutSec
	updates, err := tg.GetUpdatesChan(u)
	if err != nil {
		return nil, errors.Wrap(err, "telegram failed to get updates")
	}

	return newAdapter(ctx, tg, updates, conf)
}

func newAdapter(ctx context.Context, tg *tgbotapi.BotAPI, updates tgbotapi.UpdatesChannel, conf Config) (*TelegramAdapter, error) {
	a := &TelegramAdapter{
		BotAPI:    tg,
		updates:   updates,
		context:   ctx,
		logger:    conf.Logger,
		callbacks: make(map[string]Callback),
		conf:      conf,
	}

	if a.logger == nil {
		a.logger = zap.NewNop()
	}

	user, err := a.BotAPI.GetMe()
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

		// handle callback queries
		if update.CallbackQuery != nil {
			q := update.CallbackQuery
			l := a.logger.With(
				zap.Int("update_id", update.UpdateID),
				zap.String("callback_query_id", q.ID),
				zap.String("callback_data", q.Data))
			l.Debug("Received callback query")
			cb, ok := a.callbacks[q.Data]
			if !ok {
				l.Error("No registered callback")
				continue
			}
			go func() {
				l.Debug("Processing callback query")
				if err := cb(formatChatID(q.Message.Chat.ID)); err != nil {
					l.Error("Failed processing callback query",
						zap.Error(err))
				}
				// TODO doing this here prevents the callback from opening a URL
				l.Debug("Answering callback query")
				if _, err := a.BotAPI.AnswerCallbackQuery(tgbotapi.CallbackConfig{
					CallbackQueryID: q.ID,
					ShowAlert:       false,
				}); err != nil {
					l.Error("Failed Answering callback query",
						zap.Error(err))
				}
			}()
			continue
		}

		// skip empty every other update type
		if update.Message == nil {
			continue
		}

		m := update.Message

		if m.IsCommand() {
			arg0, argStr := m.Command(), m.CommandArguments()
			args := strings.Split(argStr, " ")

			a.logger.Debug("Received command",
				zap.Int("update_id", update.UpdateID),
				zap.String("command", arg0),
				zap.Strings("args", args),
			)
			brain.Emit(ReceiveCommandEvent{
				Arg0: arg0,
				Args: args,
				From: m.From,
				Chat: m.Chat,
				Data: m,
			})
			continue
		}

		a.logger.Debug("Received message",
			zap.Int("update_id", update.UpdateID),
			zap.Int("message_id", m.MessageID),
		)
		brain.Emit(joe.ReceiveMessageEvent{
			Text:     strings.TrimSpace(m.Text),
			Channel:  formatChatID(m.Chat.ID),
			AuthorID: strconv.Itoa(m.From.ID),
			Data:     m,
		})
	}
}

// Send sends a messages (formatted using Config.ParseMode) via the
// Telegram BotAPI
func (a *TelegramAdapter) Send(txt, chatIDString string) error {
	chatID, err := strconv.ParseInt(chatIDString, 10, 64)
	if err != nil {
		return errors.Wrap(err, "failed to parse chat id")
	}

	a.logger.Info("Sending message to chat",
		zap.Int64("chat_id", chatID),
	)

	m := tgbotapi.NewMessage(chatID, txt)
	m.ParseMode = a.conf.ParseMode
	_, err = a.BotAPI.Send(m)
	return err
}

// Close finishes processing updates
func (a *TelegramAdapter) Close() error {
	a.updates.Clear()
	return nil
}
