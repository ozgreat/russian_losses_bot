package discord

import (
	"context"
	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/log"
	"os"
	"russian_losses/pkg/bot/discord/command"
	"russian_losses/pkg/bot/discord/util"
	"russian_losses/pkg/db"
	l "russian_losses/pkg/losses"
)

var dsBot *Bot = nil

type Bot struct {
	client bot.Client
}

func newDiscordBot(token string) (*Bot, error) {
	client, createError := disgo.New(
		token,
		bot.WithGatewayConfigOpts(
			gateway.WithIntents(
				gateway.IntentGuildMessages,
				gateway.IntentMessageContent,
			),
		),

		bot.WithEventListenerFunc(onEvent),
		bot.WithEventListeners(
			command.GetStatListener(),
			command.GetDailyListener(),
		),
	)

	if createError != nil {
		return nil, createError
	}

	commands := []discord.ApplicationCommandCreate{
		command.GetStatCommand(),
		command.GetDailyCommand(),
	}

	_, regErr := client.Rest().SetGlobalCommands(client.ApplicationID(), commands)

	if regErr != nil {
		return nil, regErr
	}

	openError := client.OpenGateway(context.TODO())
	if openError != nil {
		log.Fatal("errors while connecting to gateway: ", openError)
		return nil, openError
	}

	return &Bot{client: client}, nil
}

func GetDiscordBot() (*Bot, error) {
	if dsBot == nil {
		var err error
		getenv := os.Getenv("discordToken")
		dsBot, err = newDiscordBot(getenv)
		return dsBot, err
	}

	return dsBot, nil
}

func (b Bot) SendStatistics(chats []db.ChatEntity, info *l.StatisticOfLoses) error {
	return command.SendStatistics(b.client, chats, info)
}

func (Bot) AddChat(chatId string) error {
	return util.AddChat(chatId)
}

func (b Bot) StopBot() {
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
		util.AddChat(message.ChannelID.String())
	}
}

func handlePermissionUpdate(event *events.GuildApplicationCommandPermissionsUpdate) {
	for _, p := range event.Permissions.Permissions {
		channelPermission, ok := p.(discord.ApplicationCommandPermissionChannel)
		if !ok {
			continue
		} else if channelPermission.Permission {
			util.AddChat(channelPermission.ChannelID.String())
		} else {
			util.RemoveChat(channelPermission.ChannelID.String())
		}
	}
}
