package bot

import (
	"russian_losses/pkg/db"
	"russian_losses/pkg/losses"
)

type IBot interface {
	// SendStatistics Sends formatted losses.StatisticOfLoses to all chats from chatIds array
	SendStatistics(chats []db.ChatEntity, info *losses.StatisticOfLoses) error
	// AddChat Adds chat to db to save it for later to scheduled static sending
	AddChat(chatId string) error
	// StopBot Stops bot, called before application stop
	StopBot()
}
