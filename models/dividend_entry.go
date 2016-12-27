package models

import "github.com/HT808s/gofinance/util"

type DividendEntry struct {
	Date      util.YearMonthDay
	Dividends float64 `json:",string"`
}
