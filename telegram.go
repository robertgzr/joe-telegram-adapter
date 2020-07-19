package telegram // import "github.com/robertgzr/joe-telegram-adapter"

import (
	"fmt"
	"strconv"
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
