package main

const applicationID = "1001563484553156120"

type response struct {
	Hotels []struct {
		Hotel []struct {
			HotelBasicInfo struct {
				HotelName           string `json:"hotelName"`
				HotelInformationURL string `json:"hotelInformationUrl"`
				HotelSpecial        string `json:"hotelSpecial"`
				HotelThumbnailURL   string `json:"hotelThumbnailUrl"`
			} `json:"hotelBasicInfo,omitempty"`
		} `json:"hotel"`
	} `json:"hotels"`
}
