package telegram

import (
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"gopkg.in/telegram-bot-api.v4"
)

type Telegram map[string]string

var messages chan string
var files chan *os.File

func Connect(creds map[string]string) Telegram {
	tg := Telegram{}
	for k, v := range creds {
		tg[k] = v
	}
	messages = make(chan string)
	files = make(chan *os.File)
	go tg.Serve()
	return tg
}

func (tg Telegram) Send(message string) {
	messages <- message
}
func (tg Telegram) SendFile(message *os.File) {
	files <- message
}

func (tg Telegram) Serve() {
	cidI, _ := strconv.Atoi(tg["chatID"])
	cid := int64(cidI)
	bot, err := tgbotapi.NewBotAPI(tg["token"])
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	// log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	// bot.Send(tgbotapi.NewMessage(cid, "Привет, я включилась."))

	for {
		select {
		case update := <-updates:
			if update.Message == nil {
				continue
			}
			// log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
			// msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
			// msg.ReplyToMessageID = update.Message.MessageID
			// bot.Send(msg)

			bot.Send(tgbotapi.NewPhotoUpload(cid, tg.GetCat()))
		// case message := <-files:
		case message := <-messages:
			bot.Send(tgbotapi.NewMessage(cid, message))
		}
	}
}

func (tg Telegram) GetCat() string {
	url := "http://thecatapi.com/api/images/get?format=src&type=gif"
	fname := "cat.png"
	out, _ := os.Create(fname)
	defer out.Close()
	resp, _ := http.Get(url)
	io.Copy(out, resp.Body)
	return fname
}
