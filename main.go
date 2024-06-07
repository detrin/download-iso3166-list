package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
	"github.com/jessevdk/go-flags"
)

type TableRow map[string]string

type Options struct {
	Mode       string `short:"m" long:"mode" description:"Set the run mode (options: fast, normal, slow)" default:"normal"`
	ShowWindow bool   `short:"s" long:"show-window" description:"Show browser window for debugging"`
	Version    bool   `short:"v" long:"version" description:"Show version information"`
	Timeout    int    `short:"t" long:"timeout" description:"Timeout for the entire scraping session in seconds" default:"60"`
	Help       bool   `short:"h" long:"help" description:"Show help message"`
}

const (
	version = "1.0.0"
	author  = "Daniel Herman"
	source  = "https://github.com/detrin/download-iso3166-list"
)

func main() {
	var opts Options

	parser := flags.NewParser(&opts, flags.Default)
	parser.Usage = "[OPTIONS]"

	_, err := parser.Parse()
	if err != nil {
		if !flags.WroteHelp(err) {
			fmt.Println(err)
		}
		os.Exit(1)
	}

	if opts.Help {
		parser.WriteHelp(os.Stdout)
		return
	}

	if opts.Version {
		fmt.Printf("Version: %s\n", version)
		fmt.Printf("Author: %s\n", author)
		fmt.Printf("Source: %s\n", source)
		return
	}

	var waitDuration time.Duration
	switch opts.Mode {
	case "fast":
		waitDuration = 1 * time.Second
	case "slow":
		waitDuration = 5 * time.Second
	case "normal":
		fallthrough
	default:
		waitDuration = 2 * time.Second
	}

	headlessFlag := !opts.ShowWindow
	allocatorOptions := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", headlessFlag),
		chromedp.Flag("disable-gpu", headlessFlag),
		chromedp.Flag("start-maximized", true),
	)

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), allocatorOptions...)
	defer allocCancel()

	// Create context with timeout
	timeoutCtx, timeoutCancel := context.WithTimeout(allocCtx, time.Duration(opts.Timeout)*time.Second)
	defer timeoutCancel()

	ctx, cancel := chromedp.NewContext(timeoutCtx, chromedp.WithLogf(log.Printf))
	defer cancel()

	var outerHTML string

	err = chromedp.Run(ctx,
		// navigate to the page
		chromedp.Navigate("https://www.iso.org/obp/ui/#search"),
		chromedp.WaitVisible(`#onetrust-pc-btn-handler`, chromedp.ByQuery),

		// handle cookies
		chromedp.Click(`#onetrust-pc-btn-handler`, chromedp.ByQuery),
		chromedp.WaitVisible(`#accept-recommended-btn-handler`, chromedp.ByQuery),
		chromedp.Click(`#accept-recommended-btn-handler`, chromedp.ByQuery),
		chromedp.Sleep(waitDuration),

		// select Country code
		chromedp.WaitVisible(`#gwt-uid-12`, chromedp.ByQuery),
		chromedp.Click(`#gwt-uid-12`, chromedp.ByQuery),

		// search button
		chromedp.WaitVisible(`#obpui-105541713 > div > div.v-customcomponent.v-widget.v-has-width.v-has-height > div > div > div:nth-child(2) > div > div > div.v-tabsheet-content.v-tabsheet-content-header > div > div > div > div > div > div:nth-child(2) > div > div.v-slot.v-slot-global-search.v-slot-light.v-slot-home-search > div > div.v-panel-content.v-panel-content-global-search.v-panel-content-light.v-panel-content-home-search.v-scrollable > div > div > div.v-slot.v-slot-go > div > span > span`, chromedp.ByQuery),
		chromedp.Click(`#obpui-105541713 > div > div.v-customcomponent.v-widget.v-has-width.v-has-height > div > div > div:nth-child(2) > div > div > div.v-tabsheet-content.v-tabsheet-content-header > div > div > div > div > div > div:nth-child(2) > div > div.v-slot.v-slot-global-search.v-slot-light.v-slot-home-search > div > div.v-panel-content.v-panel-content-global-search.v-panel-content-light.v-panel-content-home-search.v-scrollable > div > div > div.v-slot.v-slot-go > div > span > span`, chromedp.ByQuery),
		chromedp.Sleep(waitDuration),

		// select 300 pages
		chromedp.SetValue(`#obpui-105541713 > div > div.v-customcomponent.v-widget.v-has-width.v-has-height > div > div > div:nth-child(2) > div > div > div.v-tabsheet-content.v-tabsheet-content-header > div > div > div > div > div > div.v-slot.v-slot-search-header > div > div:nth-child(5) > div:nth-child(3) > div > select`, "8", chromedp.ByQuery),
		chromedp.Sleep(waitDuration),

		// wait for table to load
		chromedp.OuterHTML(`#obpui-105541713 > div > div.v-customcomponent.v-widget.v-has-width.v-has-height > div > div > div:nth-child(2) > div > div > div.v-tabsheet-content.v-tabsheet-content-header > div > div > div > div > div > div.v-slot.v-slot-borderless > div > div.v-panel-content.v-panel-content-borderless.v-scrollable > div > div > div.v-slot.v-slot-search-result-layout > div > div:nth-child(2) > div.v-grid.v-widget.country-code.v-grid-country-code.v-has-width > div.v-grid-tablewrapper > table`, &outerHTML, chromedp.ByQuery),
	)

	if err != nil {
		log.Fatal(err)
	}

	// parse the HTML using goquery
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(outerHTML))
	if err != nil {
		log.Fatal(err)
	}

	var jsonData []TableRow
	var headers []string

	// extract table headers
	doc.Find("thead tr th").Each(func(i int, s *goquery.Selection) {
		headers = append(headers, s.Text())
	})

	// extract table rows and cells, using headers as keys
	doc.Find("tbody tr").Each(func(i int, tr *goquery.Selection) {
		row := make(TableRow)
		tr.Find("td").Each(func(j int, td *goquery.Selection) {
			if j < len(headers) {
				row[headers[j]] = td.Text()
			}
		})
		jsonData = append(jsonData, row)
	})

	// convert jsonData to JSON
	jsonOutput, err := json.MarshalIndent(jsonData, "", "    ")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(jsonOutput))
}
