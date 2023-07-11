package bot

import (
	"context"
	"github.com/disgoorg/disgo/gateway"
	"os"
	"russian_losses/pkg/db"
	"russian_losses/pkg/losses"
	"strconv"
	"strings"

	"github.com/disgoorg/log"

	"github.com/disgoorg/disgo/events"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/snowflake/v2"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
)

var dsBot *DiscordBot = nil

type DiscordBot struct {
	client bot.Client
}

func newDiscordBot(token string) (*DiscordBot, error) {
	client, createError := disgo.New(
		token,
		bot.WithGatewayConfigOpts(
			gateway.WithIntents(
				gateway.IntentGuildMessages,
				gateway.IntentMessageContent,
			),
		),
		bot.WithEventListenerFunc(onEvent),
	)

	if createError != nil {
		return nil, createError
	}

	openError := client.OpenGateway(context.TODO())
	if openError != nil {
		log.Fatal("errors while connecting to gateway: ", openError)
		return nil, openError
	}

	return &DiscordBot{client: client}, nil
}

func GetDiscordBot() (*DiscordBot, error) {
	if dsBot == nil {
		var err error
		dsBot, err = newDiscordBot(os.Getenv("discordToken"))
		return dsBot, err
	}

	return dsBot, nil
}

func (b DiscordBot) SendStatistics(chats []db.ChatEntity, info *losses.StatisticOfLoses) error {
	for _, chat := range chats {
		if chat.BotPlatform != db.Discord {
			continue
		}

		iChatId, parseError := strconv.ParseInt(chat.ChatId, 10, 64)

		if parseError != nil {
			return parseError
		}

		go func() {
			message := info.ToMessage()

			// Fix for discord specific MarkDown
			message = strings.ReplaceAll(message, "*", "**")

			_, err := b.client.Rest().CreateMessage(
				snowflake.ID(iChatId),
				discord.NewMessageCreateBuilder().SetContent(message).Build(),
			)
			if err != nil {
				log.Error(err)
			}
		}()
	}

	return nil
}

func (DiscordBot) AddChat(chatId string) error {
	service, err := db.GetDbService()
	if err != nil {
		return err
	}

	return service.InsertChatId(chatId, db.Discord)
}

func (b DiscordBot) StopBot() {
	b.client.Close(context.TODO())
}

func onEvent(event bot.Event) {
	messageEvent, ok := event.(*events.MessageCreate)
	if ok && messageEvent.Message.Type == discord.MessageTypeUserJoin {
		handleJoinChannelEvent(messageEvent)
	}
	permissionEvent, ok := event.(*events.GuildApplicationCommandPermissionsUpdate)
	if ok {
		handlePermissionUpdate(permissionEvent)
	}
}

func handleJoinChannelEvent(event *events.MessageCreate) {
	message := event.Message
	if message.Member.User.ID == 1121874147803418644 {
		addChannelToDb(message.ChannelID.String())
	}
}

func handlePermissionUpdate(event *events.GuildApplicationCommandPermissionsUpdate) {
	for _, p := range event.Permissions.Permissions {
		channelPermission, ok := p.(discord.ApplicationCommandPermissionChannel)
		if !ok {
			continue
		} else if channelPermission.Permission {
			addChannelToDb(channelPermission.ChannelID.String())
		} else {
			removeChannelFromDb(channelPermission.ChannelID.String())
		}
	}
}

func addChannelToDb(channelId string) {
	discordBot, getBotErr := GetDiscordBot()
	if getBotErr != nil {
		log.Error(getBotErr)
	}

	addChatErr := discordBot.AddChat(channelId)
	if addChatErr != nil {
		log.Error(getBotErr)
	}
}

func removeChannelFromDb(channelId string) {
	service, dbErr := db.GetDbService()
	if dbErr != nil {
		log.Error(dbErr)
	}

	removeErr := service.RemoveChat(channelId, db.Discord)
	if removeErr != nil {
		log.Error(removeErr)
	}
}
