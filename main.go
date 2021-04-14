package main

import (
	"flag"
	"log"
	"os"
	"os/exec"
	"time"

	aw "github.com/deanishe/awgo"
	"go.deanishe.net/fuzzy"
)

var (
	cacheName = "results.json" // Filename of cached results
	// maxResults  = 200                // Number of results sent to Alfred
	maxCacheAge = 1440 * time.Minute // How long to cache results for

	// Command-line flags
	doDownload bool
	query      string

	// Workflow
	sopts []fuzzy.Option
	wf    *aw.Workflow
)

func init() {
	flag.BoolVar(&doDownload, "download", false, "retrieve list of results from gobyexample.com")

	// Set some custom fuzzy search options
	sopts = []fuzzy.Option{
		fuzzy.AdjacencyBonus(10.0),
		fuzzy.LeadingLetterPenalty(-0.1),
		fuzzy.MaxLeadingLetterPenalty(-3.0),
		fuzzy.UnmatchedLetterPenalty(-0.5),
	}
	wf = aw.New(aw.HelpURL("https://github.com/alfonmga/gobyexample-alfred-workflow"),
		// aw.MaxResults(maxResults),
		aw.SortOptions(sopts...))
}

func run() {
	wf.Args()
	flag.Parse()

	if args := flag.Args(); len(args) > 0 {
		query = args[0]
	}

	if doDownload {
		wf.Configure(aw.TextErrors(true))
		log.Printf("[main] fetching results...")
		data, err := fetchGobyexampleResults()
		if err != nil {
			wf.FatalError(err)
		}
		log.Printf("%v\n", data.toJSON())
		if err := wf.Cache.StoreJSON(cacheName, data.toJSON()); err != nil {
			wf.FatalError(err)
		}
		log.Printf("[main] fetched results!")
		return
	}

	log.Printf("[main] query=%s", query)

	var resultsRawJSON string
	if wf.Cache.Exists(cacheName) {
		if err := wf.Cache.LoadJSON(cacheName, &resultsRawJSON); err != nil {
			wf.FatalError(err)
		}
	}
	log.Printf("[main] cached results raw json %v", resultsRawJSON)
	gobyexampleData := unmarshalGobyexampleDatafromJSON([]byte(resultsRawJSON))

	if wf.Cache.Expired(cacheName, maxCacheAge) {
		wf.Rerun(0.3)
		if !wf.IsRunning("download") {
			cmd := exec.Command(os.Args[0], "-download")
			if err := wf.RunInBackground("download", cmd); err != nil {
				wf.FatalError(err)
			}
		} else {
			log.Printf("download job already running.")
		}
		if len(gobyexampleData.SectionsList) == 0 {
			wf.NewItem("Downloading gobyexample.com results…").
				Icon(aw.IconInfo)
			wf.SendFeedback()
			return
		}
	}

	for _, r := range gobyexampleData.SectionsList {
		it := wf.NewItem(r.Title).
			Arg(r.Url).
			UID(r.Url).
			Valid(true)
		it.Var("action", "url")
	}

	if query != "" {
		res := wf.Filter(query)
		log.Printf("[main] %d/%d gobyexample results match %q", len(res), len(gobyexampleData.SectionsList), query)
	}

	wf.WarnEmpty("No results found", "Try a different query?")
	wf.SendFeedback()
}

func main() {
	wf.Run(run)
}
