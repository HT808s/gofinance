package models

import "github.com/HT808s/gofinance/util"

type HistEntry struct {
	Date     util.YearMonthDay `json:"Date"`
	Open     float64           `json:"Open,string"`
	Close    float64           `json:"Close,string"`
	AdjClose float64           `json:"AdjClose,string"`
	High     float64           `json:"High,string"`
	Low      float64           `json:"Low,string"`
	Volume   int64             `json:"Volume,string"`
}
