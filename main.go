package main

import (
	"context"
	"crypto/tls"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"runtime/debug"
	"strconv"
	"strings"

	// change from 3rd party packages to golang packages
	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
	any "github.com/clbanning/anyxml"
	"github.com/dlclark/regexp2"
	// "golang.org/x/net/proxy"
)

var (
	config     *Config
	proxyIndex = 0
)

const (
	settingsConfig = "settings.json"
	scrapingConfig = "scraping.json"
	outputFile     = "output.json"
	fileName       = "scrapedata"
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

type Config struct {
	JavaScript    bool
	Proxy         bool
	ProxyLists    []string
	RotatingProxy bool
	Export        string
}

// To function properly, a lot of memory is needed to clean up files.
func clearCache() {
	// temp files
	os.RemoveAll(os.TempDir())
	debug.FreeOSMemory()
}

func readSettingsJSON() {
	// open the file and read the file
	data, err := ioutil.ReadFile(settingsConfig)
	// define data struture, json data
	var settings Config
	err = json.Unmarshal(data, &settings)
	// log any errors
	if err != nil {
		log.Println(err)
	}
	// set config to settings
	config = &settings
}

func readSiteMap() *Scraping {
	// open the file and read the file
	data, err := ioutil.ReadFile(scrapingConfig)
	// define data struture, json data
	var scrape Scraping
	err = json.Unmarshal(data, &scrape)
	// log any errors
	if err != nil {
		log.Println(err)
	}
	// return a value
	return &scrape
}

// SelectorText get data text for html tag
func SelectorText(doc *goquery.Document, selector *Selectors) []string {
	// Find the review items
	// fmt.Println(selector.Selector)
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
	// Find the review items
	// fmt.Println(selector.Selector)
	var links []string
	doc.Find(selector.Selector).EachWithBreak(func(i int, s *goquery.Selection) bool {
		href, ok := s.Attr("href")
		if !ok {
			fmt.Println("HREF has not been found")
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
	// Find the review items
	// fmt.Println(selector.Selector)
	var links []string
	doc.Find(selector.Selector).EachWithBreak(func(i int, s *goquery.Selection) bool {
		href, ok := s.Attr(selector.ExtractAttribute)
		if !ok {
			fmt.Println("HREF has not been found")
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
						fmt.Println("HREF has not been found")
					}
					elementoutput[elementSelector.ID] = resultText
				} else if elementSelector.Type == "SelectorLink" {
					resultText, ok := s.Find(elementSelector.Selector).Attr("href")
					if !ok {
						fmt.Println("HREF has not been found")
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
	// Find the review items
	// fmt.Println(selector.Selector)
	var srcs []string
	doc.Find(selector.Selector).EachWithBreak(func(i int, s *goquery.Selection) bool {
		src, ok := s.Attr("src")
		if !ok {
			fmt.Println("HREF has not been found")
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

		proxyUrl, _ := url.Parse(proxyString)

		transport = &http.Transport{
			TLSClientConfig: tls,
			Proxy:           http.ProxyURL(proxyUrl),
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


//save JSON format
func saveJSONFormat(data map[string]interface{}, format string) {

	outputFile := fmt.Sprintf("%s.%s", fileName, format)

	_, err := os.Create(outputFile)

	if err != nil {
		log.Fatal(err)
	}

	json, err := json.MarshalIndent(data, "", " ")

	if err != nil {
		log.Fatal(err)
	}

	_ = ioutil.WriteFile(outputFile, json, 0644)
}

//save xml format
func saveXMLFormat(data map[string]interface{}, format string) {

	outputFile := fmt.Sprintf("%s.%s", fileName, format)

	_, err := os.Create(outputFile)

	if err != nil {
		log.Fatal(err)
	}

	xml, err := any.XmlIndent(data, "", "   ")

	if err != nil {
		log.Fatal(err)
	}

	_ = ioutil.WriteFile(outputFile, xml, 0644)
}

// save CSV saveCsvFormat
func saveCSVFormat(data map[string]interface{}, format string) {

	var result [][]string

	outputFile := fmt.Sprintf("%s.%s", fileName, format)

	File, _ := os.Create(outputFile)

	w := csv.NewWriter(File)

	for _, v := range data {

		if rec, ok := v.(map[string]interface{}); ok {
			for key, val := range rec {
				var res []string
				res = append(res, key)
				res = append(res, fmt.Sprint(val))
				result = append(result, res)
			}
		}
	}
	w.WriteAll(result)
	w.Flush()

}

func saveScrapeData(data map[string]interface{}) {

	switch strings.ToLower(config.Export) {
	case "json":
		saveJSONFormat(data, "json")
	case "xml":
		saveXMLFormat(data, "xml")
	case "csv":
		saveCSVFormat(data, "csv")

	}
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
			// fmt.Printf("urlrange = %s ", urlRange)
			if urlRange != nil {
				val2 := strings.Replace(urlLink, fmt.Sprintf("%s", urlRange), "", -2)
				// val2 := fmt.Sprintf("%s", val)
				// fmt.Println(val2)
				urlRange2 := fmt.Sprintf("%s", urlRange)
				for _, charc := range []string{"[", "]"} {
					urlRange2 = strings.Replace(urlRange2, charc, "", -2)
				}
				rang := strings.Split(urlRange2, "-")
				// fmt.Println(rang)
				// using ParseInt method
				int1, _ := strconv.ParseInt(rang[0], 10, 64)
				int2, _ := strconv.ParseInt(rang[1], 10, 64)
				// create an anonymous inner function
				// keyword "go" starts a goroutine

				for x := int1; x <= int2; x++ {
					c <- fmt.Sprintf("%s%d", val2, x)
				}
				// close(c) sets the status of the channel c to false
				// and is needed by the for/range loop to end
				// close(c)

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

func scraper(siteMap *Scraping, parent string) map[string]interface{} {

	output := make(map[string]interface{})
	// urlLength := len(siteMap.StartURL)
	fc := getURL(siteMap.StartURL)
	// for i := 0; i < urlLength; i++ {
	if fc != nil {
		for startURL := range fc {
			// startURL := siteMap.StartURL[i]
			linkOutput := make(map[string]interface{})
			fmt.Println("URL:", startURL)
			for _, selector := range siteMap.Selectors {
				if parent == selector.ParentSelectors[0] {
					var doc *goquery.Document

					if config.JavaScript {
						doc = emulateURL(startURL)
					} else {
						doc = crawlURL(startURL)
					}

					if doc == nil {
						continue
					}

					if selector.Type == "SelectorText" {
						resultText := SelectorText(doc, &selector)
						// fmt.Printf("text resultText = %v", resultText)
						if len(resultText) != 0 {
							if len(resultText) == 1 {
								linkOutput[selector.ID] = resultText[0]
							} else {
								linkOutput[selector.ID] = resultText
							}
						}
					} else if selector.Type == "SelectorLink" {
						links := SelectorLink(doc, &selector, startURL)
						// fmt.Printf("Links = %v", links)
						if HasElem(selector.ParentSelectors, selector.ID) {
							for _, link := range links {
								if !HasElem(siteMap.StartURL, link) {
									siteMap.StartURL = append(siteMap.StartURL, link)
								}
							}

						} else {
							childSelector := getChildSelector(&selector)
							if childSelector == true {
								linkOutput[selector.ID] = links
							} else {
								newSiteMap := getSiteMap(links, &selector)
								// fmt.Printf("newSiteMap = %+v", newSiteMap)
								result := scraper(newSiteMap, selector.ID)
								// fmt.Printf("result = %+v", result)
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
						resultText := SelectorElement(doc, &selector, startURL)
						linkOutput[selector.ID] = resultText
					} else if selector.Type == "SelectorTable" {
						resultText := SelectorTable(doc, &selector)
						linkOutput[selector.ID] = resultText
					}
				}
			}
			//fmt.Printf("linkoutput = %v", linkOutput)
			if len(linkOutput) != 0 {
				if parent == "_root" {
					out, err := ioutil.ReadFile(outputFile)
					if err != nil {
						fmt.Printf("Error while reading %s file\n", outputFile)
						os.Exit(1)
					}

					var data map[string]interface{}
					err = json.Unmarshal(out, &data)
					if err != nil {
						fmt.Printf("Failed to unmarshal %s file\n", outputFile)
						os.Exit(1)
					}
					data[startURL] = linkOutput
					saveScrapeData(data)
				} else {
					output[startURL] = linkOutput
				}
			}
		}
	}
	return output
}

func main() {
	clearCache()
	_ = ioutil.WriteFile(outputFile, []byte("{}"), 0644)
	siteMap := readSiteMap()
	readSettingsJSON()

	_ = scraper(siteMap, "_root")
}