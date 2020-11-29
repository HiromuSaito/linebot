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

const backMsg = "位置情報を送信してください。"
const errorMsg = "エラーが発生しました。"
const apiURL = "https://app.rakuten.co.jp/services/api/Travel/SimpleHotelSearch/20170426?format=json&latitude=%s&longitude=%s&searchRadius=3&datumType=1&applicationId=%s"

func main() {
	http.HandleFunc("/callback", callback)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func callback(w http.ResponseWriter, r *http.Request) {
	//チャネル作成時に取得したChannel secret及びChannel access tokenを引数に渡す
	bot, err := linebot.New(config.SECRET, config.TOKEN)
	if err != nil {
		log.Fatal(err)
	}
	//http.Requestを*linebot.Eventにパースする。
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
		//リクエストのイベントがメッセージの受信かどうか
		if event.Type == linebot.EventTypeMessage {
			//受信したメッセージの種類による分岐
			switch event.Message.(type) {
			case *linebot.LocationMessage: //位置情報を受信した場合
				hotelInfoBack(bot, event)
			default: //位置情報以外を受信した場合
				_, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(backMsg)).Do()
				if err != nil {
					log.Print(err)
				}
			}
		}
	}
}

func hotelInfoBack(bot *linebot.Client, e *linebot.Event) {
	msg := e.Message.(*linebot.LocationMessage)

	//受け取った位置情報から、緯度経度を取得する
	lat := strconv.FormatFloat(msg.Latitude, 'f', 2, 64)
	lng := strconv.FormatFloat(msg.Longitude, 'f', 2, 64)

	//ホテル情報を取得する
	replyMsg, couldGetInfo := getHotelInfo(lat, lng)
	//取得に失敗した場合は、定型文（エラーが発生しました。）を返す。
	if !couldGetInfo {
		_, err := bot.ReplyMessage(e.ReplyToken, linebot.NewTextMessage(errorMsg)).Do()
		if err != nil {
			log.Print(err)
		}
	}

	//応答するカルーセルテンプレートを作成する
	res := linebot.NewTemplateMessage(
		"ホテル一覧",
		linebot.NewCarouselTemplate(replyMsg...).WithImageOptions("rectangle", "cover"),
	)
	//応答を返す
	_, err := bot.ReplyMessage(e.ReplyToken, res).Do()
	if err != nil {
		log.Print(err)
	}
}

func getHotelInfo(lat, lng string) ([]*linebot.CarouselColumn, bool) {
	url := fmt.Sprintf(apiURL, lat, lng, config.API_ID)
	r, err := http.Get(url)
	if err != nil {
		return nil, false
	}
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, false
	}

	var res response
	//構造体にパースする
	if err = json.Unmarshal(body, &res); err != nil {
		return nil, false
	}

	var ccs []*linebot.CarouselColumn

	//カルーセルカラムの作成
	for index, hotel := range res.Hotels {
		if index == 10 {
			break
		}
		cc := linebot.NewCarouselColumn(
			hotel.Hotel[0].HotelBasicInfo.HotelThumbnailURL,
			cutOutCharacters(hotel.Hotel[0].HotelBasicInfo.HotelName, 40),
			cutOutCharacters(hotel.Hotel[0].HotelBasicInfo.HotelSpecial, 60),
			linebot.NewURIAction("楽天トラベルで開く", hotel.Hotel[0].HotelBasicInfo.HotelInformationURL),
		).WithImageOptions("#FFFFFF")
		ccs = append(ccs, cc)

	}
	return ccs, true
}

//cutOutCharacters 先頭から指定文字数だけを切りだす("abcde",3)　→　"abc"
func cutOutCharacters(s string, count int) string {
	if utf8.RuneCountInString(s) > count {
		return string([]rune(s)[:count])
	}
	return s
}
