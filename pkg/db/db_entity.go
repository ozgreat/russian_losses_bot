package db

type (
	Platform   string
	StatFormat string
	BotLang    string
)

const (
	Discord  Platform = "DISCORD"
	Telegram Platform = "TELEGRAM"

	UA BotLang = "UA" // TODO multi-language support feature

	Text      StatFormat = "TEXT"
	Image     StatFormat = "IMG" // TODO add image info
	TextImage StatFormat = "TEXT/IMG"
)

type ChatEntity struct {
	ChatId            string
	BotPlatform       Platform
	Lang              BotLang
	Format            StatFormat
	DailyNotification bool
}

type dbQuery func(chat ChatEntity, prevErr error) error

type FormatError struct {
	Msg string
}

func (e FormatError) Error() string {
	return e.Msg
}
