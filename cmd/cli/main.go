package main

import (
	"fmt"

	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/HT808s/gofinance"
	"github.com/HT808s/gofinance/sources/bloomberg"
	"github.com/HT808s/gofinance/sources/yahoofinance"
)

// Variables linked during build
var (
	Version    string
	Githash    string
	Buildstamp string
	AppName    string
	SourcesMap map[string]gofinance.Source

	symbols = kingpin.Flag("symbols", "ticker symbols to analyze").Short('s').Required().Strings()
	source  = kingpin.Flag("source", "bloomberg || yahoo_yql || yahoo_cvs").
		Default("yahoo_yql").Enum("bloomberg", "yahoo_yql", "yahoo_cvs")
)

func init() {
	SourcesMap = map[string]gofinance.Source{
		"bloomberg": bloomberg.New(),
		"yahoo_yql": yahoofinance.NewYql(),
		"yahoo_cvs": yahoofinance.NewCvs()}
}

func main() {
	kingpin.Version(fmt.Sprintf("%s [%s] Commit: %s Buildtstamp: %s",
		AppName, Version, Githash, Buildstamp))
	kingpin.Parse()

	app := App{Symbols: *symbols, Src: SourcesMap[*source]}
	app.Run()
}
