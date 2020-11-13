package main

import (
	"context"
	"crypto/tls"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
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

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
	"github.com/dlclark/regexp2"
)

var (
	config     *Config
	outputFile = "output"
)

const (
	settingsConfig = "settings.json"
	scrapingConfig = "sitemap.json"
	logFile        = "logs.log"
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
	Gui        bool
	Log        bool
	JavaScript bool
	Workers    int
	Export     string
	UserAgents []string
	Captcha    []string
	Proxy      []string
}

// WorkerJob struct defination
type WorkerJob struct {
	startURL string
	parent   string
	siteMap  *Scraping
	// doc        *goquery.Document
	linkOutput map[string]interface{}
}

// All the device memory is needed, so all the temp files are removed.
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
		fmt.Println("Error: Temporary files can't be deleted.")
	}
}

// Reading the settings json
// Future Update: Merge all of the Jsons into one.
func readSettingsJSON() {
	data, err := ioutil.ReadFile(settingsConfig)
	var settings Config
	err = json.Unmarshal(data, &settings)
	config = &settings
	if err != nil {
		if config.Log {
			file, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
			defer file.Close()
			log.SetOutput(file)
			log.Println(err)
			os.Exit(0)
		}
		log.Println(err)
		os.Exit(0)
	}
}

// read the scraping json
func readSiteMap() *Scraping {
	data, err := ioutil.ReadFile(scrapingConfig)
	var scrape Scraping
	err = json.Unmarshal(data, &scrape)
	if err != nil {
		if config.Log {
			file, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
			defer file.Close()
			log.SetOutput(file)
			log.Println(err)
			os.Exit(0)
		}
		log.Println(err)
		os.Exit(0)
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

func crawlURL(href, userAgent string) *goquery.Document {
	var transport *http.Transport

	tls := &tls.Config{
		InsecureSkipVerify: false,
	}
	// if proxy is set use for transport
	if len(config.Proxy) > 0 {

		proxyString := config.Proxy[0]

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

	req, err := http.NewRequest(http.MethodGet, href, nil)
	if err != nil {
		if config.Log {
			file, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
			defer file.Close()
			log.SetOutput(file)
			log.Println(err)
			os.Exit(0)
		}
		log.Println(err)
		os.Exit(0)
	}

	if len(userAgent) > 0 {
		req.Header.Set("User-Agent", userAgent)
	}

	response, err := netClient.Do(req)
	if err != nil {
		if config.Log {
			file, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
			defer file.Close()
			log.SetOutput(file)
			log.Println(err)
			os.Exit(0)
		}
		log.Println(err)
		os.Exit(0)
	}

	// bodyBytes, err := ioutil.ReadAll(response.Body)
	// fmt.Println(string(bodyBytes))

	defer response.Body.Close()

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(response.Body)
	return doc
}

func toFixedURL(href, baseURL string) string {
	uri, err := url.Parse(href)

	base, err := url.Parse(baseURL)
	if err != nil {
		if config.Log {
			file, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
			defer file.Close()
			log.SetOutput(file)
			log.Println(err)
			os.Exit(0)
		}
		log.Println(err)
		os.Exit(0)
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

func emulateURL(url, userAgent string) *goquery.Document {
	var opts []func(*chromedp.ExecAllocator)

	if len(config.Proxy) > 0 {

		proxyString := config.Proxy[0]
		proxyServer := chromedp.ProxyServer(proxyString)
		// fmt.Println(proxyServer)
		opts = append(chromedp.DefaultExecAllocatorOptions[:], proxyServer)
	} else {
		opts = append(chromedp.DefaultExecAllocatorOptions[:])
	}

	if len(userAgent) > 0 {
		opts = append(opts, chromedp.UserAgent(userAgent))
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

	r := strings.NewReader(body)

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		if config.Log {
			file, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
			defer file.Close()
			log.SetOutput(file)
			log.Println(err)
			os.Exit(0)
		}
		log.Println(err)
		os.Exit(0)
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
	userAgents := config.UserAgents

	if len(userAgents) == 0 {
		userAgents = append(userAgents, "")
	}

	for count := 0; count < len(userAgents); count++ {

		userAgent := userAgents[count]

		for job := range jobs {

			var doc *goquery.Document

			if config.JavaScript {
				doc = emulateURL(job.startURL, userAgent)
			} else {
				doc = crawlURL(job.startURL, userAgent)
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
	for x := 1; x <= config.Workers; x++ {
		wg.Add(1)
		go worker(x, jobs, results, &wg)
	}

	// go saveDataToFile(results, outputChannel)
	go func() {
		fc := getURL(siteMap.StartURL)
		if fc != nil {
			for startURL := range fc {
				// fmt.Println("URL:", startURL)
				if !validURL(startURL) {
					continue
				}

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
						if config.Log {
							file, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
							defer file.Close()
							log.SetOutput(file)
							log.Println(err)
							os.Exit(0)
						}
						log.Println(err)
						os.Exit(0)
					}

					var data map[string]interface{}
					err = json.Unmarshal(out, &data)
					data[job.startURL] = job.linkOutput

					switch config.Export {
					case "XML":
						output, err := xml.MarshalIndent(data, "", " ")
						if err != nil {
							if config.Log {
								file, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
								defer file.Close()
								log.SetOutput(file)
								log.Println(err)
								os.Exit(0)
							}
							log.Println(err)
							os.Exit(0)
						}

						_ = ioutil.WriteFile(outputFile, output, 0644)

					case "CSV":
						csvFile, err := os.OpenFile(outputFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
						if err != nil {
							if config.Log {
								file, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
								defer file.Close()
								log.SetOutput(file)
								log.Println(err)
								os.Exit(0)
							}
							log.Println(err)
							os.Exit(0)
						}

						csvWriter := csv.NewWriter(csvFile)
						rows := [][]string{}

						for i, v := range data {
							rows = append(rows, []string{i, fmt.Sprint(v)})
						}

						for _, row := range rows {
							_ = csvWriter.Write(row)
						}

						csvWriter.Flush()

						csvFile.Close()
					case "JSON":
						output, err := json.MarshalIndent(data, "", " ")
						if err != nil {
							if config.Log {
								file, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
								defer file.Close()
								log.SetOutput(file)
								log.Println(err)
								os.Exit(0)
							}
							log.Println(err)
							os.Exit(0)
						}
						_ = ioutil.WriteFile(outputFile, output, 0644)
					default:
						fmt.Println("Error: Please choose a output format.")
					}
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

func validURL(uri string) bool {
	_, err := url.ParseRequestURI(uri)
	if err != nil {
		if config.Log {
			file, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
			defer file.Close()
			log.SetOutput(file)
			log.Println(err)
		}
		log.Println(err)
		return false
	}

	return true
}

//outputResult set output file name and temp output file based on settings.json
func outputResult() {
	userFormat := strings.ToLower(config.Export)
	var allowedFormat = map[string]bool{
		"CSV":  true,
		"XML":  true,
		"JSON": true,
	}

	if allowedFormat[userFormat] {
		outputFile = fmt.Sprintf("output.%s", userFormat)
	}

	_ = ioutil.WriteFile(outputFile, []byte("{}"), 0644)
}

func main() {
	clearCache()
	siteMap := readSiteMap()
	readSettingsJSON()
	outputResult()
	_ = scraper(siteMap, "_root")
}
