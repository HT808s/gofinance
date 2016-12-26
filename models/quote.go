package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"time"

	. "github.com/HT808s/gofinance/term"
	"github.com/fatih/color"
)

const (
	ErrTplNotSupported = "source '%s' does not support action '%s'"
)

type Quotes []*Quote

type Quote struct {
	Symbol   string /* e.g.: VEUR.AS, Vanguard dev. europe on Amsterdam */
	Name     string
	Exchange string

	/* last actualization of the results */
	Updated time.Time

	/* volume */
	Volume         int64 /* outstanding shares */
	AvgDailyVolume int64 /* avg amount of shares traded */

	/* dividend & related */
	PeRatio          float64   /* Price / EPS */
	EarningsPerShare float64   /* (net income - spec.dividends) / avg.  outstanding shares */
	DividendPerShare float64   /* total (non-special) dividend payout / total amount of shares */
	DividendYield    float64   /* annual div. per share / price per share */
	DividendExDate   time.Time /* last dividend payout date */

	/* price & derived */
	Bid, Ask              float64
	Open, PreviousClose   float64
	LastTradePrice        float64
	Change, ChangePercent float64

	DayLow, DayHigh   float64
	YearLow, YearHigh float64

	Ma50, Ma200 float64 /* 200- and 50-day moving average */
}

/* will try to calculate the dividend payout ratio, if possible,
 * otherwise returns 0 */
func (q *Quote) DivPayoutRatio() float64 {
	/* total dividends / net income (same period, mostly 1Y):
	 * TODO: implement this (first implement historical data
	 * aggregation) */

	/* DPS / EPS */
	if q.DividendPerShare != 0 && q.EarningsPerShare != 0 {
		return q.DividendPerShare / q.EarningsPerShare
	}
	return 0
}

func (q *Quote) GetPrice() float64 {
	if q.Ask != 0 {
		return q.Ask
	}
	return q.LastTradePrice
}

// SharesToBuy gives you the number of shares to buy if you want
//  * the transaction cost to be less than a certain percentage
func (q *Quote) SharesToBuy(txCost, desiredTxCostPerc float64) float64 {
	return math.Ceil((txCost - desiredTxCostPerc*txCost) /
		(desiredTxCostPerc * q.GetPrice()))
}

// True if the Quote is increasing
// False otherwise
func (q *Quote) IsIncreasing() bool {
	return q.LastTradePrice >= q.PreviousClose
}

func (q *Quote) VariationValue() float64 {
	return q.LastTradePrice - q.PreviousClose
}

func (q *Quote) VariationValuePercent() float64 {
	return q.VariationValue() / q.PreviousClose * 100
}

func (q *Quote) WouldBuyOrSell() interface{} {
	return q.PreviousClose > q.Ma200
}

func (q *Quote) PrettyDisplay(txCost, desiredTxCostPerc float64) string {
	var buffer bytes.Buffer
	maxBidAskSpreadPerc := 0.01
	minDivYield := 0.025
	amountOfsharesForLowTxCost := q.SharesToBuy(txCost, desiredTxCostPerc)

	buffer.WriteString(fmt.Sprintf("\n%v (%v), %v %v %v\n",
		q.Name, q.Symbol,
		Binary(fmt.Sprintf("%+.2f", q.VariationValue()), q.IsIncreasing()),
		Binary(fmt.Sprintf("%+.2f%%", q.VariationValuePercent()), q.IsIncreasing()),
		Binary(Arrow(q.IsIncreasing()), q.IsIncreasing())))

	if q.Bid != 0 && q.Ask != 0 {
		bidAskSpreadPerc := (q.Ask - q.Bid) / q.Bid
		bidAskPrint := Binaryfp(bidAskSpreadPerc*100, bidAskSpreadPerc < maxBidAskSpreadPerc)
		buffer.WriteString(fmt.Sprintf("bid/ask: %v, spread: %v (%v)\n",
			color.MagentaString("%.2f/%.2f", q.Bid, q.Ask),
			color.MagentaString("%.2f", q.Ask-q.Bid),
			bidAskPrint))
		if bidAskSpreadPerc < maxBidAskSpreadPerc {
			buffer.WriteString(
				fmt.Sprintf("if you want to buy this stock, place a %v at about %v\n",
					color.GreenString("limit order"), color.GreenString("%f", (q.Ask+q.Bid)/2)))
		} else {
			buffer.WriteString(
				fmt.Sprintf("%v the spread of this stock is rather high",
					color.RedString("CAUTION:")))
		}
	}

	buffer.WriteString(fmt.Sprintf("prev_close/open/last_trade: %v\n",
		color.MagentaString("%.2f/%.2f/%.2f",
			q.PreviousClose, q.Open, q.LastTradePrice)))
	buffer.WriteString(fmt.Sprintf("day low/high: %v\n",
		color.MagentaString("%.2f/%.2f (%.2f)",
			q.DayLow, q.DayHigh, q.DayHigh-q.DayLow)))
	buffer.WriteString(fmt.Sprintf("year low/high: %v\n",
		color.MagentaString("%.2f/%.2f (%.2f)",
			q.YearLow, q.YearHigh, q.YearHigh-q.YearLow)))
	buffer.WriteString(fmt.Sprintf("moving avg. 50/200: %v\n",
		color.MagentaString("%.2f/%.2f", q.Ma50, q.Ma200)))
	divYield := Binaryfp(q.DividendYield*100, q.DividendYield > minDivYield)
	buffer.WriteString(
		fmt.Sprintf("last ex-dividend: %v, div. per share: %v, div. yield: %v\n"+
			"earnings per share: %v, dividend payout ratio: %v\n",
			q.DividendExDate.Format("02/01"),
			color.MagentaString("%.2f", q.DividendPerShare),
			divYield,
			q.EarningsPerShare,
			color.MagentaString("%.2f", q.DivPayoutRatio())))
	buffer.WriteString(
		fmt.Sprintf("You would need to buy %v (€ %v) shares of this stock to reach a transaction cost below %v%%\n",
			color.GreenString("%f", amountOfsharesForLowTxCost),
			color.GreenString("%f", amountOfsharesForLowTxCost*q.GetPrice()),
			desiredTxCostPerc*100))
	if q.PeRatio != 0 {
		buffer.WriteString(color.MagentaString("The P/E-ratio is %v, ", q.PeRatio))
		switch {
		case 0 <= q.PeRatio && q.PeRatio <= 10:
			underv := color.GreenString("undervalued")
			decline := color.RedString("market thinks its earnings are going to decline")
			above := color.GreenString("above the historic trend for this company")
			buffer.WriteString(
				fmt.Sprintf("this stock is either %v or the %v, either that or the companies earnings are %v\n",
					underv, decline, above))
		case 11 <= q.PeRatio && q.PeRatio <= 17:
			buffer.WriteString(
				fmt.Sprintf("this usually represents a %s.\n",
					color.GreenString("fair value")))
		case 18 <= q.PeRatio && q.PeRatio <= 25:
			overv := color.RedString("overvalued")
			incrlast := color.GreenString("earnings have increased since the last earnings call")
			increxp := color.GreenString("earnings expected to increase substantially in the future")
			buffer.WriteString(
				fmt.Sprintf("either the stock is %v or the %v figure was published. The stock may also be a growth stock with %v.\n",
					overv, incrlast, increxp))
		case 26 <= q.PeRatio:
			bubble := color.RedString("bubble")
			earnings := color.GreenString("very high expected earnings")
			low := color.RedString("this years earnings have been exceptionally low (unlikely)")
			buffer.WriteString(
				fmt.Sprintf("Either we're in a %v, or the company has %v, or %v\n",
					bubble, earnings, low))
		}
	}

	if q.Ma200 != 0 {
		buffer.WriteString("Richie Rich thinks this is a ")
		if q.WouldBuyOrSell().(bool) {
			buffer.WriteString(color.GreenString("BUY position\n"))
		} else {
			buffer.WriteString(color.RedString("SELL position\n"))
		}
	}
	buffer.WriteString("======================")
	return buffer.String()
}

func (q *Quote) JSON() string {
	b, _ := json.Marshal(q)
	return string(b)
}
