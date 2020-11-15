package main

import (
	"log"
	"net/http"

	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/linebot/src/config"
)

func main() {
	http.HandleFunc("/callback", callback)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

const messageBack = "位置情報を送信してください。"

func callback(w http.ResponseWriter, r *http.Request) {
	bot, err := linebot.New(config.SECRET, config.TOKEN)
	if err != nil {
		log.Fatal(err)
	}
	events, err := bot.ParseRequest(r)
	if err != nil {
		if err == linebot.ErrInvalidSignature {
			w.WriteHeader(400) //Bad Request
		} else {
			w.WriteHeader(500) //Internal Server Error
		}
		return
	}

	for _, event := range events {
		if event.Type == linebot.EventTypeMessage {
			switch event.Message.(type) {
			//テキストを受信した場合
			case *linebot.TextMessage:
				_, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(messageBack)).Do()
				if err != nil {
					log.Print(err)
				}
			}
		}
	}
}
