package bot

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	hbot "github.com/whyrusleeping/hellabot"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	llog "gopkg.in/inconshreveable/log15.v2"

	"github.com/adzip-kadum/irc-calc/log"
	"github.com/adzip-kadum/irc-calc/postgres"
	"github.com/adzip-kadum/irc-calc/repository"
)

type Config struct {
	Channel   string   `yaml:"channel"`
	Nickname  string   `yaml:"nick"`
	Addresses []string `yaml:"addresses"`
	Encoding  string   `yaml:"encoding"`
}

type Bot struct {
	conf    Config
	repo    *repository.CalcsRepository
	encoder *encoding.Encoder
	decoder *encoding.Decoder
}

type (
	encoderMaker func() *encoding.Encoder
	decoderMaker func() *encoding.Decoder
)

var (
	encoders = map[string]encoderMaker{
		"windows-1251": func() *encoding.Encoder { return charmap.Windows1251.NewEncoder() },
		"koi8-r":       func() *encoding.Encoder { return charmap.KOI8R.NewEncoder() },
	}
	decoders = map[string]decoderMaker{
		"windows-1251": func() *encoding.Decoder { return charmap.Windows1251.NewDecoder() },
		"koi8-r":       func() *encoding.Decoder { return charmap.KOI8R.NewDecoder() },
	}
)

func NewBot(conf Config, pool *postgres.PgxPool) (*Bot, error) {
	bot := &Bot{
		conf: conf,
		repo: repository.NewCalcsRepository(pool),
	}
	encoder := encoders[conf.Encoding]
	decoder := decoders[conf.Encoding]

	if encoder == nil || decoder == nil {
		return nil, errors.Errorf("Unknown encoder or decoder for %q", conf.Encoding)
	}

	bot.encoder = encoder()
	bot.decoder = decoder()

	return bot, nil
}

const prefix = "!calc "

func (b *Bot) Start() error {
	log.Info("starting bot", log.Any("config", b.conf))

	hijackSession := func(bot *hbot.Bot) {
		bot.HijackSession = true
	}
	channels := func(bot *hbot.Bot) {
		bot.Channels = []string{b.conf.Channel}
	}
	bot, err := hbot.NewBot(b.conf.Addresses[0], b.conf.Nickname, hijackSession, channels)
	if err != nil {
		return err
	}
	trigger := hbot.Trigger{
		Condition: func(bot *hbot.Bot, m *hbot.Message) bool {
			return strings.HasPrefix(m.Content, prefix)
		},
		Action: func(irc *hbot.Bot, m *hbot.Message) bool {
			decodedFrom, err := b.decoder.String(m.From)
			if err != nil {
				irc.Reply(m, fmt.Sprintf("ERROR: %s", err))
				return false
			}
			decodedContent, err := b.decoder.String(m.Content)
			if err != nil {
				irc.Reply(m, fmt.Sprintf("ERROR: %s", err))
				return false
			}
			calc, err := b.getOrSetCalc(decodedFrom, m.TimeStamp, decodedContent)
			if err != nil {
				irc.Reply(m, fmt.Sprintf("ERROR: %s", err))
				return false
			}
			irc.Reply(m, calc)
			return true
		}}
	bot.AddTrigger(trigger)
	//bot.Logger.SetHandler(llog.LvlFilterHandler())
	bot.Logger.SetHandler(llog.StreamHandler(os.Stdout, llog.JsonFormat()))
	go bot.Run()

	return nil
}

func (b *Bot) Stop() error {
	return nil
}

var (
	spaces   = regexp.MustCompile(`\s+`)
	hasIndex = regexp.MustCompile(`\[(\d+)\]$`)
)

func (b *Bot) getOrSetCalc(from string, when time.Time, data string) (string, error) {
	parts := strings.SplitN(data, "=", 2)
	if len(parts) > 1 {
		return b.setCalc(from, when, parts)
	}
	key := spaces.ReplaceAllString(data, " ")
	index := hasIndex.FindAllStringSubmatch(key, 1)
	if len(index) > 0 {
		key = key[:len(key)-len(index[0][0])]
	}
	key = strings.TrimSpace(key[len(prefix):])
	calcs, err := b.repo.GetCalcs(context.Background(), repository.GetCalcsParams{
		Channel: b.conf.Channel,
		Key:     key,
	})
	if err != nil {
		return fmt.Sprintf("ERROR %s", err), nil
	}
	if len(calcs) == 0 {
		return fmt.Sprintf("there is no calcs with %q", key), nil
	}
	if len(index) > 0 {
		return b.getCalcByIndex(index[0][1], calcs)
	}
	return b.getCalcByIndex("0", calcs)
}

func (b *Bot) getCalcByIndex(index string, calcs []repository.IrcCalc) (string, error) {
	num, err := strconv.ParseUint(index, 10, 64)
	if err != nil {
		return "", err
	}
	if int(num) > len(calcs)-1 {
		return "", errors.Errorf("calc index %d out of range, max %d", num, len(calcs)-1)
	}
	c := calcs[num]
	encodedKey, err := b.encoder.String(c.Key)
	if err != nil {
		return "", err
	}
	encodedContent, err := b.encoder.String(c.Content)
	if err != nil {
		return "", err
	}
	encodedBy, err := b.encoder.String(c.By)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s = %s [%s, %s]", encodedKey, encodedContent, encodedBy, formatTime(c.When)), nil
}

func (b *Bot) setCalc(by string, when time.Time, data []string) (string, error) {
	key := spaces.ReplaceAllString(data[0], " ")
	key = strings.TrimSpace(key[len(prefix):])
	content := spaces.ReplaceAllString(data[1], " ")
	content = strings.TrimSpace(content)
	params := repository.AddCalcParams{
		Channel: b.conf.Channel,
		Key:     key,
		By:      by,
		When:    when.UTC(),
		Content: content,
	}
	if _, err := b.repo.AddCalc(context.Background(), params); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s = %s [%s, %s]", key, content, by, formatTime(when)), nil
}

func formatTime(t time.Time) string {
	return t.UTC().Format(time.RFC850)
}
