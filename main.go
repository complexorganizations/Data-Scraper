package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	// "encoding/xml"
	// "encoding/csv"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"

	// change from 3rd party packages to golang packages
	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
	"github.com/dlclark/regexp2"
	// "golang.org/x/net/proxy"
)

var (
	config          *Config
	proxyIndex      = 0
	numberOfWorkers = 10
)

const (
	settingsConfig = "settings.json"
	scrapingConfig = "scraping.json"
	outputFile     = "output.json"
	//logFile        = "logging.log"
)

// Selectors is struct to Marshal selector
type Selectors struct {
	ID               string
	Type             string
	ParentSelectors  []string
	Selector         string
	Multiple         bool
	Regex            string
	Delay            int
	ExtractAttribute string
}

// Scraping is struct to Marshal scraping file
type Scraping struct {
	StartURL  []string
	ID        string `json:"_id,omitempty"`
	Selectors []Selectors
}

// Config setting struct
type Config struct {
	JavaScript    bool
	Log           bool
	Proxy         bool
	ProxyLists    []string
	RotatingProxy bool
	Export        string
}

// WorkerJob struct defination
type WorkerJob struct {
	startURL string
	parent   string
	siteMap  *Scraping
	// doc        *goquery.Document
	linkOutput map[string]interface{}
}

// To function properly, a lot of memory is needed to clean up files.
func clearCache() {
	operatingSystem := runtime.GOOS
	switch operatingSystem {
	case "windows":
		os.RemoveAll(os.TempDir())
		debug.FreeOSMemory()
	case "darwin":
		os.RemoveAll(os.TempDir())
		debug.FreeOSMemory()
	case "linux":
		os.RemoveAll(os.TempDir())
		debug.FreeOSMemory()
	default:
		fmt.Println("Error: The func is not supported in your OS.")
	}
}

// read the settings json
func readSettingsJSON() {
	data, err := ioutil.ReadFile(settingsConfig)
	var settings Config
	err = json.Unmarshal(data, &settings)
	if err != nil {
		log.Println(err)
	}
	config = &settings
}

// read the scraping json
func readSiteMap() *Scraping {
	data, err := ioutil.ReadFile(scrapingConfig)
	var scrape Scraping
	err = json.Unmarshal(data, &scrape)
	if err != nil {
		log.Println(err)
	}
	return &scrape
}

// SelectorText get data text for html tag
func SelectorText(doc *goquery.Document, selector *Selectors) []string {
	var text []string
	var matchText *regexp2.Match
	doc.Find(selector.Selector).EachWithBreak(func(i int, s *goquery.Selection) bool {
		if selector.Regex != "" {
			re := regexp2.MustCompile(selector.Regex, 0)
			if matchText, _ = re.FindStringMatch(s.Text()); matchText != nil {
				text = append(text, strings.TrimSpace(matchText.String()))
			} else {
				text = append(text, strings.TrimSpace(s.Text()))
			}
		} else {
			text = append(text, strings.TrimSpace(s.Text()))
		}
		if selector.Multiple == false {
			return false
		}
		return true
	})
	return text
}

// SelectorLink get data href for html tag
func SelectorLink(doc *goquery.Document, selector *Selectors, baseURL string) []string {
	var links []string
	doc.Find(selector.Selector).EachWithBreak(func(i int, s *goquery.Selection) bool {
		href, ok := s.Attr("href")
		if !ok {
			fmt.Println("Error: HREF has not been found.")
		}

		links = append(links, toFixedURL(href, baseURL))
		if selector.Multiple == false {
			return false
		}
		return true
	})
	return links
}

// SelectorElementAttribute get define attribute for html tag
func SelectorElementAttribute(doc *goquery.Document, selector *Selectors) []string {
	var links []string
	doc.Find(selector.Selector).EachWithBreak(func(i int, s *goquery.Selection) bool {
		href, ok := s.Attr(selector.ExtractAttribute)
		if !ok {
			fmt.Println("Error: HREF has not been found.")
		}

		links = append(links, href)
		if selector.Multiple == false {
			return false
		}
		return true
	})
	return links
}

// SelectorElement get child element of html selected element
func SelectorElement(doc *goquery.Document, selector *Selectors, startURL string) []interface{} {
	baseSiteMap := readSiteMap()
	var elementoutputList []interface{}
	doc.Find(selector.Selector).EachWithBreak(func(i int, s *goquery.Selection) bool {
		elementoutput := make(map[string]interface{})
		for _, elementSelector := range baseSiteMap.Selectors {
			if selector.ID == elementSelector.ParentSelectors[0] {
				if elementSelector.Type == "SelectorText" {
					// resultText := SelectorText(s, elementSelector)
					resultText := s.Find(elementSelector.Selector).Text()
					elementoutput[elementSelector.ID] = resultText
				} else if elementSelector.Type == "SelectorImage" {
					resultText, ok := s.Find(elementSelector.Selector).Attr("src")
					if !ok {
						fmt.Println("Error: HREF has not been found.")
					}
					elementoutput[elementSelector.ID] = resultText
				} else if elementSelector.Type == "SelectorLink" {
					resultText, ok := s.Find(elementSelector.Selector).Attr("href")
					if !ok {
						fmt.Println("Error: HREF has not been found.")
					}
					elementoutput[elementSelector.ID] = resultText
				}
			}
		}
		if len(elementoutput) != 0 {
			elementoutputList = append(elementoutputList, elementoutput)
		}
		if selector.Multiple == false {
			return false
		}
		return true

	})
	return elementoutputList
}

// SelectorImage get src of Image for html tag
func SelectorImage(doc *goquery.Document, selector *Selectors) []string {
	var srcs []string
	doc.Find(selector.Selector).EachWithBreak(func(i int, s *goquery.Selection) bool {
		src, ok := s.Attr("src")
		if !ok {
			fmt.Println("Error: HREF has not been found.")
		}
		srcs = append(srcs, src)
		if selector.Multiple == false {
			return false
		}
		return true
	})
	return srcs
}

// SelectorTable get header and row data of table
func SelectorTable(doc *goquery.Document, selector *Selectors) map[string]interface{} {
	var headings, row []string
	var rows [][]string
	table := make(map[string]interface{})
	doc.Find(selector.Selector).Each(func(index int, tablehtml *goquery.Selection) {
		tablehtml.Find("tr").Each(func(indextr int, rowhtml *goquery.Selection) {
			rowhtml.Find("th").Each(func(indexth int, tableheading *goquery.Selection) {
				headings = append(headings, tableheading.Text())
			})
			rowhtml.Find("td").Each(func(indexth int, tablecell *goquery.Selection) {
				row = append(row, tablecell.Text())
			})
			if len(row) != 0 {
				rows = append(rows, row)
				row = nil
			}
		})
	})
	table["header"] = headings
	table["rows"] = rows

	return table
}

func crawlURL(href string) *goquery.Document {
	var transport *http.Transport

	tls := &tls.Config{
		InsecureSkipVerify: false,
	}
	// if proxy is set use for transport
	if config.Proxy {
		var proxyString string

		if config.RotatingProxy {
			if proxyIndex <= len(config.ProxyLists) {
				proxyString = config.ProxyLists[proxyIndex]
				proxyIndex = proxyIndex + 1
			} else {
				proxyString = config.ProxyLists[0]
				proxyIndex = 0
			}
		} else {
			proxyString = config.ProxyLists[0]
		}

		proxyURL, _ := url.Parse(proxyString)

		transport = &http.Transport{
			TLSClientConfig: tls,
			Proxy:           http.ProxyURL(proxyURL),
		}
	} else {
		transport = &http.Transport{
			TLSClientConfig: tls,
		}
	}

	netClient := &http.Client{
		Transport: transport,
	}

	response, err := netClient.Get(href)
	if err != nil {
		log.Println(err)
		return nil
	}
	defer response.Body.Close()

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		log.Println(err)
	}

	return doc
}

func toFixedURL(href, baseURL string) string {
	uri, err := url.Parse(href)
	if err != nil {
		return ""
	}

	base, err := url.Parse(baseURL)
	if err != nil {
		return ""
	}

	toFixedURI := base.ResolveReference(uri)

	return toFixedURI.String()
}

func getSiteMap(startURL []string, selector *Selectors) *Scraping {

	baseSiteMap := readSiteMap()
	newSiteMap := new(Scraping)
	newSiteMap.ID = selector.ID
	newSiteMap.StartURL = startURL
	newSiteMap.Selectors = baseSiteMap.Selectors
	return newSiteMap
}

func getChildSelector(selector *Selectors) bool {
	baseSiteMap := readSiteMap()
	var count int = 0
	for _, childSelector := range baseSiteMap.Selectors {
		if selector.ID == childSelector.ParentSelectors[0] {
			count++
		}
	}
	if count == 0 {
		return true
	}
	return false
}

// HasElem check element is present or not in parsed list
func HasElem(s interface{}, elem interface{}) bool {
	arrV := reflect.ValueOf(s)
	if arrV.Kind() == reflect.Slice {
		for i := 0; i < arrV.Len(); i++ {

			// XXX - panics if slice element points to an unexported struct field
			// see https://golang.org/pkg/reflect/#Value.Interface
			if arrV.Index(i).Interface() == elem {
				return true
			}
		}
	}

	return false
}

func emulateURL(url string) *goquery.Document {
	var proxyString string
	var opts []func(*chromedp.ExecAllocator)

	if config.Proxy {
		if config.RotatingProxy {
			if proxyIndex <= len(config.ProxyLists) {
				proxyString = config.ProxyLists[proxyIndex]
				proxyIndex = proxyIndex + 1
			} else {
				proxyString = config.ProxyLists[0]
				proxyIndex = 0
			}
		} else {
			proxyString = config.ProxyLists[0]
		}

		opts = append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.ProxyServer(proxyString),
		)
	} else {
		opts = append(chromedp.DefaultExecAllocatorOptions[:])
	}

	// create context
	bctx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, _ := chromedp.NewContext(bctx)
	defer cancel()

	var err error

	// run task list
	var body string

	err = chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.InnerHTML(`body`, &body, chromedp.NodeVisible, chromedp.ByQuery),
	)

	if err != nil {
		log.Println(err)
		return nil
	}

	r := strings.NewReader(body)

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		log.Println(err)
	}

	return doc
}

// generator using a channel and a goroutine
func getURL(urls []string) <-chan string {

	// create a channel
	c := make(chan string)
	go func() {
		re := regexp2.MustCompile(`(\[\d{1,10}-\d{1,10}\]$)`, 0)

		for _, urlLink := range urls {
			urlRange, _ := re.FindStringMatch(urlLink)
			if urlRange != nil {
				val2 := strings.Replace(urlLink, fmt.Sprintf("%s", urlRange), "", -2)
				urlRange2 := fmt.Sprintf("%s", urlRange)
				for _, charc := range []string{"[", "]"} {
					urlRange2 = strings.Replace(urlRange2, charc, "", -2)
				}
				rang := strings.Split(urlRange2, "-")
				// using ParseInt method
				int1, _ := strconv.ParseInt(rang[0], 10, 64)
				int2, _ := strconv.ParseInt(rang[1], 10, 64)

				// Send url in channel
				for x := int1; x <= int2; x++ {
					c <- fmt.Sprintf("%s%d", val2, x)
				}

			} else {
				c <- urlLink
			}
		}
		// close(c) sets the status of the channel c to false
		// and is needed by the for/range loop to end
		close(c)
	}()
	return c
}

func worker(workerID int, jobs <-chan WorkerJob, results chan<- WorkerJob, wg *sync.WaitGroup) {
	defer wg.Done()
	// fmt.Printf("Worker %d started\n", workerID)
	for job := range jobs {

		var doc *goquery.Document

		if config.JavaScript {
			doc = emulateURL(job.startURL)
		} else {
			doc = crawlURL(job.startURL)
		}

		if doc == nil {
			continue
		}
		fmt.Println("URL:", job.startURL)
		linkOutput := make(map[string]interface{})
		for _, selector := range job.siteMap.Selectors {
			if job.parent == selector.ParentSelectors[0] {
				if selector.Type == "SelectorText" {
					resultText := SelectorText(doc, &selector)
					if len(resultText) != 0 {
						if len(resultText) == 1 {
							linkOutput[selector.ID] = resultText[0]
						} else {
							linkOutput[selector.ID] = resultText
						}
					}
				} else if selector.Type == "SelectorLink" {
					links := SelectorLink(doc, &selector, job.startURL)
					// fmt.Printf("Links = %v", links)
					if HasElem(selector.ParentSelectors, selector.ID) {
						for _, link := range links {
							if !HasElem(job.siteMap.StartURL, link) {
								job.siteMap.StartURL = append(job.siteMap.StartURL, link)
							}
						}
					} else {
						childSelector := getChildSelector(&selector)
						if childSelector == true {
							linkOutput[selector.ID] = links
						} else {
							newSiteMap := getSiteMap(links, &selector)
							result := scraper(newSiteMap, selector.ID)
							linkOutput[selector.ID] = result
						}
					}
				} else if selector.Type == "SelectorElementAttribute" {
					resultText := SelectorElementAttribute(doc, &selector)
					linkOutput[selector.ID] = resultText
				} else if selector.Type == "SelectorImage" {
					resultText := SelectorImage(doc, &selector)
					if len(resultText) != 0 {
						if len(resultText) == 1 {
							linkOutput[selector.ID] = resultText[0]
						} else {
							linkOutput[selector.ID] = resultText
						}
					}
				} else if selector.Type == "SelectorElement" {
					resultText := SelectorElement(doc, &selector, job.startURL)
					linkOutput[selector.ID] = resultText
				} else if selector.Type == "SelectorTable" {
					resultText := SelectorTable(doc, &selector)
					linkOutput[selector.ID] = resultText
				}
			}
		}
		job.linkOutput = linkOutput
		results <- job
	}
	//fmt.Println("Stopped Worker:", workerID)
}

func scraper(siteMap *Scraping, parent string) map[string]interface{} {
	output := make(map[string]interface{})
	var wg sync.WaitGroup

	jobs := make(chan WorkerJob, 10)
	results := make(chan WorkerJob, 10)
	outputChannel := make(chan map[string]interface{})
	// 3 Workers
	for x := 1; x <= numberOfWorkers; x++ {
		wg.Add(1)
		go worker(x, jobs, results, &wg)
	}

	// go saveDataToFile(results, outputChannel)
	go func() {
		fc := getURL(siteMap.StartURL)
		if fc != nil {
			for startURL := range fc {
				// fmt.Println("URL:", startURL)

				workerjob := WorkerJob{
					parent:   parent,
					startURL: startURL,
					siteMap:  siteMap,
				}

				jobs <- workerjob
			}
			close(jobs)
		}
	}()

	go func() {
		pageOutput := make(map[string]interface{})
		for job := range results {
			if len(job.linkOutput) != 0 {
				if job.parent == "_root" {
					out, err := ioutil.ReadFile(outputFile)
					if err != nil {
						fmt.Printf("Error: Cant read %s file.\n", outputFile)
						os.Exit(1)
					}
					var data map[string]interface{}
					err = json.Unmarshal(out, &data)
					if err != nil {
						fmt.Printf("Error: Failed to unmarshal %s file.\n", outputFile)
						os.Exit(1)
					}
					data[job.startURL] = job.linkOutput
					file, err := json.MarshalIndent(data, "", " ")
					if err != nil {
						fmt.Println(err.Error())
						os.Exit(1)
					}

					_ = ioutil.WriteFile(outputFile, file, 0644)
				} else {
					pageOutput[job.startURL] = job.linkOutput
				}
			}
		}
		outputChannel <- pageOutput
	}()

	// close(jobs)
	// close(outputChannel)
	wg.Wait()
	close(results)
	output = <-outputChannel
	return output
}

func main() {
	clearCache()
	_ = ioutil.WriteFile(outputFile, []byte("{}"), 0644)
	siteMap := readSiteMap()
	readSettingsJSON()

	_ = scraper(siteMap, "_root")
}
