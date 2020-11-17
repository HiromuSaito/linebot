package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"unicode/utf8"

	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/linebot/src/config"
)

const messageBack = "位置情報を送信してください。"
const errorMsg = "エラーが発生しました。"

func main() {
	http.HandleFunc("/callback", callback)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

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
			//位置情報を受信した場合
			case *linebot.LocationMessage:
				hotelInfoBack(bot, event)
			//位置情報以外を受信した場合
			default:
				_, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(messageBack)).Do()
				if err != nil {
					log.Print(err)
				}
			}
		}
	}
}

func hotelInfoBack(bot *linebot.Client, e *linebot.Event) {
	msg := e.Message.(*linebot.LocationMessage)

	lat := strconv.FormatFloat(msg.Latitude, 'f', 2, 64)
	lng := strconv.FormatFloat(msg.Longitude, 'f', 2, 64)

	replyMsg, couldGetInfo := getHotelInfo(lat, lng)
	if !couldGetInfo {
		_, err := bot.ReplyMessage(e.ReplyToken, linebot.NewTextMessage(errorMsg)).Do()
		if err != nil {
			log.Print(err)
		}
	}

	res := linebot.NewTemplateMessage(
		"ホテル一覧",
		linebot.NewCarouselTemplate(replyMsg...).WithImageOptions("rectangle", "cover"),
	)

	_, err := bot.ReplyMessage(e.ReplyToken, res).Do()
	if err != nil {
		log.Print(err)
	}
}

func getHotelInfo(lat, lng string) ([]*linebot.CarouselColumn, bool) {
	url := fmt.Sprintf("https://app.rakuten.co.jp/services/api/Travel/SimpleHotelSearch/20170426?format=json&latitude=%s&longitude=%s&searchRadius=3&datumType=1&applicationId=%s", lat, lng, config.API_ID)
	res, err := http.Get(url)
	if err != nil {
		return nil, false
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, false
	}

	var data response
	if err = json.Unmarshal(body, &data); err != nil {
		return nil, false
	}

	var ccs []*linebot.CarouselColumn

	for _, hotel := range data.Hotels {
		special := hotel.Hotel[0].HotelBasicInfo.HotelSpecial
		if utf8.RuneCountInString(special) > 60 {
			special = string([]rune(special)[:60])
		}

		cc := linebot.NewCarouselColumn(
			hotel.Hotel[0].HotelBasicInfo.HotelThumbnailURL,
			hotel.Hotel[0].HotelBasicInfo.HotelName,
			special,
			linebot.NewURIAction("楽天トラベルで開く", hotel.Hotel[0].HotelBasicInfo.HotelInformationURL),
		).WithImageOptions("#FFFFFF")
		ccs = append(ccs, cc)
	}
	return ccs, true
}
