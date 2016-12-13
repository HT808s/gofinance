package main

import (
	"fmt"
	"math"
	"os"
	"time"

	"github.com/HT808s/gofinance/bloomberg"
	"github.com/HT808s/gofinance/fquery"
	"github.com/HT808s/gofinance/util"
	"github.com/HT808s/gofinance/yahoofinance"
	"github.com/fatih/color"
)

const (
	CONFIG_SUBPATH = ".gofinance"
	DB_FILENAME    = "gofinance.db"
)

func ConfigDir() string {
	if path := os.Getenv("GOFINANCE_DIR"); path != "" {
		return path
	} else {
		return util.Home() + "/" + CONFIG_SUBPATH
	}
}

func DbPath() string {
	if path := os.Getenv("GOFINANCE_DB"); path != "" {
		return path
	} else {
		return ConfigDir() + "/" + DB_FILENAME
	}
}

func main() {
	fmt.Printf("welcome to gofinance %v.%v.%v\n", MAJ_VERSION, MIN_VERSION, MIC_VERSION)

	var src fquery.Source
	// s := yahoofinance.NewCvs()
	src = bloomberg.New()
	src = yahoofinance.NewYql()

	// symbols := []string{
	// 	"VEUR.AS",
	// 	"VJPN.AS",
	// 	"VHYL.AS",
	// 	"AAPL",
	// 	"APC.F",
	// 	"GSZ.PA",
	// 	"COFB.BR",
	// 	"BEFB.BR",
	// 	"GIMB.BR",
	// 	"ELI.BR",
	// 	"DELB.BR",
	// 	"BELG.BR",
	// 	"TNET.BR",
	// }

	symbols := []string{
		"VEUR.AS",
		"VFEM.AS",
		"BELG.BR",
		"UMI.BR",
		"SOLB.BR",
		"KBC.BR",
		"DIE.BR",
		"DL.BR",
		"BEKB.BR",
		"ACKB.BR",
		"ABI.BR",
		"EURUSD=X",
	}

	bloomberg.VERBOSITY = 2

	// cache, err := newCache(src)
	// if err != nil {
	// 	fmt.Printf("WARNING: could not initialize cache (%v), going to use pure source\n\t", err)
	// } else {
	// 	cache.SetQuoteExpiry(5 * time.Minute)
	// 	defer cache.Close()
	// 	src = cache
	// }

	// divhist(src)
	// hist(src, symbols...)
	calc(src, symbols...)
}

func divhist(src fquery.Source) {
	res, err := src.DividendHist([]string{"BELG.BR", "TNET.BR"})
	if err != nil {
		fmt.Println("gofinance: could not fetch history, ", err)
		return
	}

	fmt.Printf("succesfully fetched %v symbols' dividend history\n", len(res))

	for symb, hist := range res {
		fmt.Println(symb)
		fmt.Println("===========")
		fmt.Println("Length:", len(hist.Dividends))
		for _, row := range hist.Dividends {
			fmt.Println("row:", row.Date.GetTime().Format("02-01-2006"), row.Dividends)
		}
	}
}

func hist(src fquery.Source, symbols ...string) {
	res, err := src.Hist(symbols)
	if err != nil {
		fmt.Println("gofinance: could not fetch history, ", err)
		return
	}

	fmt.Println("Printing history for symbols:", symbols)
	for symb, hist := range res {
		fmt.Println(symb)
		fmt.Println("===========")
		fmt.Println("Length:", len(hist.Entries))
		for _, row := range hist.Entries {
			t := time.Time(row.Date)
			fmt.Printf("%v: %v (%v)\n", t.Format("02/01/2006"), row.Close, row)
		}
		fmt.Println("Moving average manual calc:", movingAverage(hist))
	}

	/* with time limits */
	// start := time.Date(2009, time.November, 1, 0, 0, 0, 0, time.UTC)
	// end := time.Date(2011, time.November, 1, 0, 0, 0, 0, time.UTC)
	// _, err = src.HistLimit([]string{"AAPL", "VEUR.AS", "VJPN.AS"}, start, end)
	// if err != nil {
	// 	fmt.Println("gofinance: could not fetch history, ", err)
	// }
}

func calc(src fquery.Source, symbols ...string) {
	fmt.Println("requesting information on individual stocks...", symbols)
	res, err := src.Quote(symbols)
	if err != nil {
		fmt.Println("gofinance: could not fetch, ", err)
		return
	}

	desiredTxCostPerc := 0.01
	txCost := 9.75
	maxBidAskSpreadPerc := 0.01
	minDivYield := 0.025

	fmt.Println()
	for _, r := range res {
		price := nvl(r.Ask, r.LastTradePrice)
		amountOfsharesForLowTxCost := sharesToBuy(price, txCost, desiredTxCostPerc)

		upDir := r.LastTradePrice >= r.PreviousClose
		upVal := r.LastTradePrice - r.PreviousClose
		upPerc := upVal / r.PreviousClose * 100
		fmt.Printf("name: %v (%v), %v %v %v (chart: %v)\n",
			r.Name, r.Symbol,
			binary(fmt.Sprintf("%+.2f", upVal), upDir),
			binary(fmt.Sprintf("%+.2f%%", upPerc), upDir),
			binary(arrow(upDir), upDir),
			yahoofinance.GenChartUrl(r.Symbol, yahoofinance.Year2, nil))

		if r.Bid != 0 && r.Ask != 0 {
			bidAskSpreadPerc := (r.Ask - r.Bid) / r.Bid
			bidAskPrint := binaryfp(bidAskSpreadPerc*100, bidAskSpreadPerc < maxBidAskSpreadPerc)
			fmt.Printf("bid/ask: %v, spread: %v (%v)\n",
				color.MagentaString("%.2f/%.2f", r.Bid, r.Ask),
				color.MagentaString("%.2f", r.Ask-r.Bid),
				bidAskPrint)
			if bidAskSpreadPerc < maxBidAskSpreadPerc {
				fmt.Printf("if you want to buy this stock, place a %v at about %v\n", color.GreenString("limit order"), color.GreenString("%f", (r.Ask+r.Bid)/2))
			} else {
				fmt.Printf("%v the spread of this stock is rather high", color.RedString("CAUTION:"))
			}
		}

		fmt.Printf("prev_close/open/last_trade: %v\n",
			color.MagentaString("%.2f/%.2f/%.2f",
				r.PreviousClose, r.Open, r.LastTradePrice))
		fmt.Printf("day low/high: %v\n",
			color.MagentaString("%.2f/%.2f (%.2f)",
				r.DayLow, r.DayHigh, r.DayHigh-r.DayLow))
		fmt.Printf("year low/high: %v\n",
			color.MagentaString("%.2f/%.2f (%.2f)",
				r.YearLow, r.YearHigh, r.YearHigh-r.YearLow))
		fmt.Printf("moving avg. 50/200: %v\n",
			color.MagentaString("%.2f/%.2f", r.Ma50, r.Ma200))
		divYield := binaryfp(r.DividendYield*100, r.DividendYield > minDivYield)
		fmt.Printf("last ex-dividend: %v, div. per share: %v, div. yield: %v\n"+
			"earnings per share: %v, dividend payout ratio: %v\n",
			r.DividendExDate.Format("02/01"),
			color.MagentaString("%.2f", r.DividendPerShare),
			divYield,
			r.EarningsPerShare,
			color.MagentaString("%.2f", r.DivPayoutRatio()))
		fmt.Printf("You would need to buy %v (€ %v) shares of this stock to reach a transaction cost below %v%%\n",
			color.GreenString("%f", amountOfsharesForLowTxCost),
			color.GreenString("%f", amountOfsharesForLowTxCost*price),
			desiredTxCostPerc*100)
		if r.PeRatio != 0 {
			// terminal.Stdout.Colorf("The P/E-ratio is @m%.2f@|, ", r.PeRatio)
			color.Magenta("The P/E-ratio is %v, ", r.PeRatio)
			switch {
			case 0 <= r.PeRatio && r.PeRatio <= 10:
				underv := color.GreenString("undervalued")
				decline := color.RedString("market thinks its earnings are going to decline")
				above := color.GreenString("above the historic trend for this company")
				fmt.Printf("this stock is either %v or the %v, either that or the companies earnings are %v\n", underv, decline, above)
			case 11 <= r.PeRatio && r.PeRatio <= 17:
				fmt.Println("this usually represents fair value.", color.GreenString("fair value"))
			case 18 <= r.PeRatio && r.PeRatio <= 25:
				overv := color.RedString("overvalued")
				incrlast := color.GreenString("earnings have increased since the last earnings call")
				increxp := color.GreenString("earnings expected to increase substantially in the future")
				fmt.Printf("either the stock is %v or the %v figure was published. The stock may also be a growth stock with %v.\n",
					overv, incrlast, increxp)
			case 26 <= r.PeRatio:
				bubble := color.RedString("bubble")
				earnings := color.GreenString("very high expected earnings")
				low := color.RedString("this years earnings have been exceptionally low (unlikely)")
				fmt.Printf("Either we're in a %v, or the company has %v, or %v\n",
					bubble, earnings, low)
			}
			fmt.Println()
		}

		if r.Ma200 != 0 {
			fmt.Print("Richie Rich thinks this is in a ")
			if wouldRichieRichBuy(r) {
				color.Green("BUY position")
			} else {
				color.Red("SELL position")
			}
		}

		fmt.Println("======================")
	}
}

/* gives you the number of shares to buy if you want
 * the transaction cost to be less than a certain percentage
 * (0.5% is fantastic, 1% is ok, for example) */
func sharesToBuy(price, txCost, desiredTxCostPerc float64) float64 {
	return math.Ceil((txCost - desiredTxCostPerc*txCost) /
		(desiredTxCostPerc * price))
}

/* you can get the moving average for 50 and 200 days
 * out of the standard stock quote, but if you need
 * something else, just point it at this function */
func movingAverage(hist fquery.Hist) float64 {
	if len(hist.Entries) == 0 {
		return 0
	}

	var sum float64 = 0
	var count float64 = 0

	for _, row := range hist.Entries {
		sum += row.Close
		count++
	}

	return sum / count
}

func wouldRichieRichBuy(res fquery.Quote) bool {
	return res.PreviousClose > res.Ma200
}

/* returns the first non-zero float */
func nvl(xs ...float64) float64 {
	for _, x := range xs {
		if x != 0 {
			return x
		}
	}

	return 0
}
