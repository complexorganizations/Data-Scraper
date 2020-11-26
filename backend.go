package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/cdproto/target"
	"github.com/chromedp/chromedp"
	"github.com/dlclark/regexp2"
)

var (
	settings settingsT
	sitemap  scraping
)

const (
	configFile = "sitemap.json"
)

type selectors struct {
	ID               string   `json:"id"`
	Type             string   `json:"type"`
	ParentSelectors  []string `json:"parentSelectors"`
	Selector         string   `json:"selector"`
	Multiple         bool     `json:"multiple"`
	Regex            string   `json:"regex"`
	Delay            int      `json:"delay"`
	ExtractAttribute string   `json:"extractAttribute"`
}

type scraping struct {
	ID        string      `json:"projectID,omitempty"`
	StartURL  []string    `json:"startURL"`
	Selectors []selectors `json:"selectors"`
}

type settingsT struct {
	Gui          bool     `json:"gui"`
	Log          bool     `json:"log"`
	LogFile      string   `json:"logFile"`
	JavaScript   bool     `json:"javaScript"`
	Workers      int      `json:"workers"`
	RateLimiting int      `json:"rateLimiting"`
	Export       string   `json:"export"`
	OutputFile   string   `json:"outputFile"`
	UserAgents   []string `json:"userAgents"`
	Captcha      string   `json:"captcha"`
	Proxy        []string `json:"proxy"`
}

type jsonType struct {
	Settings settingsT `json:"settings"`
	Sitemap  scraping  `json:"sitemap"`
}

type workerJob struct {
	startURL   string
	parent     string
	siteMap    *scraping
	linkOutput map[string]interface{}
}

type websiteData map[string]interface{}

type xmlMapEntry struct {
	XMLName xml.Name
	Value   interface{} `xml:",chardata"`
}

type audioPostBody struct {
	Audio  audioPostAudio    `json:"audio"`
	Config recognitionConfig `json:"config"`
}

type audioPostAudio struct {
	Content string `json:"content"`
}

type speechRecognitionResponse struct {
	Result []speechRecognitionAlternativeResult `json:"results"`
}

type speechRecognitionAlternativeResult struct {
	Alternatives []speechRecognitionAlternative `json:"alternatives"`
	ChannelTag   int                            `json:"channelTag"`
}

type speechRecognitionAlternative struct {
	Transcript string     `json:"transcript"`
	Confidence float64    `json:"confidence"`
	Words      []wordInfo `json:"words"`
}

type wordInfo struct {
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
	Word      string `json:"word"`
}

type recognitionConfig struct {
	LanguageCode string `json:"languageCode"`
	Model        string `json:"model"`
}

func (m websiteData) MarshalXML(e *xml.Encoder, start xml.StartElement) error {

	if len(m) == 0 {
		return nil
	}

	err := e.EncodeToken(start)
	if err != nil {
		return err
	}

	for k, v := range m {
		e.Encode(xmlMapEntry{XMLName: xml.Name{Local: k}, Value: v})
	}

	return e.EncodeToken(start.End())
}

func clearCache() {
	operatingSystem := runtime.GOOS
	var err error
	switch operatingSystem {
	case "windows":
		err = os.RemoveAll(os.TempDir())
	case "darwin":
		err = os.RemoveAll(os.TempDir())
	case "linux":
		err = os.RemoveAll(os.TempDir())
	default:
		fmt.Println("Error: Temporary files can't be deleted.")
	}

	if err != nil {
		frontendLog(err)
	}
}

func logErrors(error error) {
	if settings.Log {
		file, err := os.OpenFile(settings.LogFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		log.SetOutput(file)
		if err != nil {
			log.SetOutput(os.Stderr)
			_, _ = fmt.Fprintf(os.Stderr, "Can't open log file: %s, printing to stderr...\n", settings.LogFile)
		}

		log.Println(error)

		if err == nil {
			err = file.Close()
			_, _ = fmt.Fprintf(os.Stderr, "Error closing log file: %s!\n", settings.LogFile)
		}
	}
}

func readJSON() {
	jsonData := jsonType{}
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		logErrors(err)
	}

	err = json.Unmarshal(data, &jsonData)
	if err != nil {
		logErrors(err)
	}

	sitemap = jsonData.Sitemap
	settings = jsonData.Settings
}

func writeJSON() {
	jsonData := jsonType{settings, sitemap}
	dataJSON, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		logErrors(err)
	}

	err = ioutil.WriteFile(configFile, dataJSON, 0644)
	if err != nil {
		logErrors(err)
	}
}

func selectorText(doc *goquery.Document, selector *selectors) []string {
	var text []string
	var matchText *regexp2.Match
	doc.Find(selector.Selector).EachWithBreak(
		func(i int, s *goquery.Selection) bool {
			if selector.Regex != "" {
				re := regexp2.MustCompile(selector.Regex, 0)
				matchText, _ = re.FindStringMatch(s.Text())
				if matchText != nil {
					text = append(text, strings.TrimSpace(matchText.String()))
				} else {
					text = append(text, strings.TrimSpace(s.Text()))
				}
			} else {
				text = append(text, strings.TrimSpace(s.Text()))
			}

			return selector.Multiple
		},
	)
	return text
}

func selectorLink(doc *goquery.Document, selector *selectors, baseURL string) []string {
	var links []string
	doc.Find(selector.Selector).EachWithBreak(
		func(i int, s *goquery.Selection) bool {
			href, errors := s.Attr("href")
			if !errors {
				fmt.Println("Error: HREF has not been found.")
			}

			links = append(links, toFixedURL(href, baseURL))

			return selector.Multiple
		},
	)
	return links
}

func selectorElementAttribute(doc *goquery.Document, selector *selectors) []string {
	var links []string
	doc.Find(selector.Selector).EachWithBreak(
		func(i int, s *goquery.Selection) bool {
			href, errors := s.Attr(selector.ExtractAttribute)
			if !errors {
				fmt.Println("Error: HREF has not been found.")
			}
			links = append(links, href)

			return selector.Multiple
		},
	)
	return links
}

func selectorElement(doc *goquery.Document, selector *selectors) []interface{} {
	baseSiteMap := sitemap
	var elementOutputList []interface{}
	doc.Find(selector.Selector).EachWithBreak(
		func(i int, s *goquery.Selection) bool {
			elementOutput := make(map[string]interface{})
			for _, elementSelector := range baseSiteMap.Selectors {
				if selector.ID == elementSelector.ParentSelectors[0] {
					if elementSelector.Type == "SelectorText" {
						resultText := s.Find(elementSelector.Selector).Text()
						elementOutput[elementSelector.ID] = resultText
					} else if elementSelector.Type == "SelectorImage" {
						resultText, errors := s.Find(elementSelector.Selector).Attr("src")
						if !errors {
							fmt.Println("Error: HREF has not been found.")
						}
						elementOutput[elementSelector.ID] = resultText
					} else if elementSelector.Type == "SelectorLink" {
						resultText, errors := s.Find(elementSelector.Selector).Attr("href")
						if !errors {
							fmt.Println("Error: HREF has not been found.")
						}
						elementOutput[elementSelector.ID] = resultText
					}
				}
			}
			if len(elementOutput) != 0 {
				elementOutputList = append(elementOutputList, elementOutput)
			}

			return selector.Multiple
		},
	)
	return elementOutputList
}

func selectorImage(doc *goquery.Document, selector *selectors) []string {
	var sources []string
	doc.Find(selector.Selector).EachWithBreak(func(i int, s *goquery.Selection) bool {
		src, errors := s.Attr("src")
		if !errors {
			fmt.Println("Error: HREF has not been found.")
		}
		sources = append(sources, src)

		return selector.Multiple
	})
	return sources
}

func selectorTable(doc *goquery.Document, selector *selectors) map[string]interface{} {
	var headings, row []string
	var rows = [][]string{}
	table := make(map[string]interface{})
	doc.Find(selector.Selector).Each(func(_ int, tableHTML *goquery.Selection) {
		tableHTML.Find("tr").Each(func(_ int, rowHTML *goquery.Selection) {
			rowHTML.Find("th").Each(func(_ int, tableHeading *goquery.Selection) {
				headings = append(headings, tableHeading.Text())
			})
			rowHTML.Find("td").Each(func(_ int, tableCell *goquery.Selection) {
				row = append(row, tableCell.Text())
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

func parseCatchAudio(url string) (string, error) {
	var speechBody speechRecognitionResponse
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)

	// Write the audio body to the buffer
	_, err = io.Copy(buf, resp.Body)

	if err != nil {
		return "", err
	}

	audioBody := &audioPostBody{
		Audio: audioPostAudio{
			Content: base64.RawURLEncoding.EncodeToString(buf.Bytes()),
		},
		Config: recognitionConfig{
			LanguageCode: "en-US",
			Model:        "video",
		},
	}

	reqBody, err := json.Marshal(audioBody)

	// pass audio into google speech api
	speechResp, err := http.Post("https://speech.googleapis.com/v1p1beta1/speech:recognize?key="+settings.Captcha, "application/json", bytes.NewBuffer(reqBody))

	if err != nil {
		return "", err
	}

	defer speechResp.Body.Close()

	err = json.NewDecoder(speechResp.Body).Decode(&speechBody)

	if err != nil {
		return "", err
	}

	return speechBody.Result[0].Alternatives[0].Transcript, nil
}

func crawlURL(href, userAgent string) *goquery.Document {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
	}

	if len(settings.Proxy) > 0 {
		proxyString := settings.Proxy[0]
		proxyURL, _ := url.Parse(proxyString)
		transport.Proxy = http.ProxyURL(proxyURL)
	}

	netClient := &http.Client{Transport: transport}
	req, err := http.NewRequest(http.MethodGet, href, nil)
	if err != nil {
		logErrors(err)
		os.Exit(1)
	}
	if len(userAgent) > 0 {
		req.Header.Set("User-Agent", userAgent)
	}
	response, err := netClient.Do(req)
	if err != nil {
		logErrors(err)
		os.Exit(1)
	}

	doc, err := goquery.NewDocumentFromReader(response.Body)

	err = response.Body.Close()

	if err != nil {
		frontendLog(err)
	}
	return doc
}

func toFixedURL(href, baseURL string) string {
	uri, err := url.Parse(href)
	if err != nil {
		logErrors(err)
		os.Exit(1)
	}
	base, err := url.Parse(baseURL)
	if err != nil {
		logErrors(err)
		os.Exit(1)
	}
	toFixedURI := base.ResolveReference(uri)
	return toFixedURI.String()
}

func getSiteMap(startURL []string, selector *selectors) *scraping {
	baseSiteMap := sitemap
	newSiteMap := new(scraping)
	newSiteMap.ID = selector.ID
	newSiteMap.StartURL = startURL
	newSiteMap.Selectors = baseSiteMap.Selectors
	return newSiteMap
}

func getChildSelector(selector *selectors) bool {
	count := 0
	for _, childSelector := range sitemap.Selectors {
		if selector.ID == childSelector.ParentSelectors[0] {
			count++
		}
	}

	return count == 0
}

func hasElement(s interface{}, elem interface{}) bool {
	arrV := reflect.ValueOf(s)
	if arrV.Kind() == reflect.Slice {
		for i := 0; i < arrV.Len(); i++ {
			if arrV.Index(i).Interface() == elem {
				return true
			}
		}
	}
	return false
}

func emulateURL(url, userAgent string) *goquery.Document {
	var opts []func(*chromedp.ExecAllocator)
	if len(settings.Proxy) > 0 {
		proxyString := settings.Proxy[0]
		proxyServer := chromedp.ProxyServer(proxyString)
		opts = append(chromedp.DefaultExecAllocatorOptions[:], proxyServer)
	} else {
		opts = append(chromedp.DefaultExecAllocatorOptions[:])
	}
	if len(userAgent) > 0 {
		opts = append(opts, chromedp.UserAgent(userAgent))
	}
	bCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, _ := chromedp.NewContext(bCtx)
	defer cancel()
	var err error
	var body string
	err = chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.InnerHTML(`body`, &body, chromedp.NodeVisible, chromedp.ByQuery),
	)
	r := strings.NewReader(body)
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		logErrors(err)
		os.Exit(1)
	}
	return doc
}

func navigateURL(url, userAgent string) *goquery.Document {
	var opts []func(*chromedp.ExecAllocator)
	if len(settings.Proxy) > 0 {
		proxyString := settings.Proxy[0]
		proxyServer := chromedp.ProxyServer(proxyString)
		opts = append(chromedp.DefaultExecAllocatorOptions[:], proxyServer)
	} else {
		opts = append(chromedp.DefaultExecAllocatorOptions[:])
	}
	if len(userAgent) > 0 {
		opts = append(opts, chromedp.UserAgent(userAgent))
	}

	bCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, cancel := chromedp.NewContext(bCtx)

	defer cancel()

	var checkboxNode *target.Info
	var challengeNode *target.Info

	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitReady("iframe", chromedp.ByQuery),
	)

	if err != nil {
		logErrors(err)
		os.Exit(0)
	}

	// need to get captcha iframe targets out
	targets, _ := chromedp.Targets(ctx)

	for _, t := range targets {
		if t.Type == "iframe" && strings.Contains(t.URL, "anchor") {
			checkboxNode = t
		}
		if t.Type == "iframe" && strings.Contains(t.URL, "bframe") {
			challengeNode = t
		}
	}
	// set context to captcha checkbox iframe
	ictx, _ := chromedp.NewContext(ctx, chromedp.WithTargetID(checkboxNode.TargetID))

	var ok bool
	var checked string

	err = chromedp.Run(
		ctx,
		chromedp.WaitVisible(`#recaptcha-anchor`, chromedp.NodeVisible),
		chromedp.Click(`#recaptcha-anchor`, chromedp.ByID),
	)

	err = chromedp.Run(
		ictx,
		chromedp.AttributeValue(`#recaptcha-anchor`, "aria-checked", &checked, &ok),
	)

	if err != nil {
		logErrors(err)
		os.Exit(0)
	}

	isCheched, _ := strconv.ParseBool(checked)

	if !isCheched {
		var audioSource *string
		ictx2, _ := chromedp.NewContext(ctx, chromedp.WithTargetID(challengeNode.TargetID))

		err = chromedp.Run(
			ictx2,
			chromedp.WaitVisible(`#recaptcha-audio-button`, chromedp.ByID),
			chromedp.Click(`#recaptcha-audio-button`, chromedp.NodeVisible),
			chromedp.WaitVisible(`#audio-response`, chromedp.ByID),
			chromedp.AttributeValue(`#audio-source`, "src", audioSource, &ok),
		)

		if err != nil {
			logErrors(err)
			os.Exit(0)
		}

		if audioSource != nil {
			text, err := parseCatchAudio(*audioSource)

			if err != nil {
				logErrors(err)
				os.Exit(0)
			}

			err = chromedp.Run(
				ictx2,
				chromedp.WaitVisible(`#audio-response`, chromedp.ByID),
				chromedp.SetValue(`#audio-response`, text, chromedp.ByID),
				chromedp.Click(`#recaptcha-verify-button`, chromedp.NodeVisible),
			)
		}
	}

	var body string
	err = chromedp.Run(ctx,
		chromedp.InnerHTML(`body`, &body, chromedp.NodeVisible, chromedp.ByQuery),
	)

	r := strings.NewReader(body)
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		logErrors(err)
		os.Exit(0)
	}

	return doc
}

func getURL(urls []string) <-chan string {
	c := make(chan string)
	go func() {
		re := regexp2.MustCompile(`(\[\d{1,10}-\d{1,10}\]$)`, 0)
		for _, urlLink := range urls {
			stringMatch, _ := re.FindStringMatch(urlLink)
			if stringMatch != nil {
				val2 := strings.Replace(urlLink, fmt.Sprintf("%s", stringMatch), "", -2)

				urlRange := fmt.Sprintf("%s", stringMatch)
				urlRange = strings.Replace(urlRange, "[", "", -2)
				urlRange = strings.Replace(urlRange, "]", "", -2)

				rang := strings.Split(urlRange, "-")
				int1, _ := strconv.ParseInt(rang[0], 10, 64)
				int2, _ := strconv.ParseInt(rang[1], 10, 64)
				for x := int1; x <= int2; x++ {
					c <- fmt.Sprintf("%s%d", val2, x)
				}
			} else {
				c <- urlLink
			}
		}
		close(c)
	}()
	return c
}

func worker(jobs <-chan workerJob, results chan<- workerJob, wg *sync.WaitGroup) {
	defer wg.Done()
	userAgents := settings.UserAgents
	if len(userAgents) == 0 {
		userAgents = append(userAgents, "")
	}
	for count := 0; count < len(userAgents); count++ {
		userAgent := userAgents[count]
		for job := range jobs {
			var doc *goquery.Document
			if settings.JavaScript {
				if settings.Captcha != "" {
					doc = navigateURL(job.startURL, userAgent)
				} else {
					doc = emulateURL(job.startURL, userAgent)
				}
			} else {
				doc = crawlURL(job.startURL, userAgent)
			}
			if doc == nil {
				continue
			}
			fmt.Println("URL:", job.startURL)
			linkOutput := make(map[string]interface{})
			for _, selector := range job.siteMap.Selectors {
				if len(selector.ParentSelectors) > 0 && job.parent == selector.ParentSelectors[0] {
					if selector.Type == "SelectorText" {
						resultText := selectorText(doc, &selector)
						if len(resultText) != 0 {
							if len(resultText) == 1 {
								linkOutput[selector.ID] = resultText[0]
							} else {
								linkOutput[selector.ID] = resultText
							}
						}
					} else if selector.Type == "SelectorLink" {
						links := selectorLink(doc, &selector, job.startURL)
						if hasElement(selector.ParentSelectors, selector.ID) {
							for _, link := range links {
								if !hasElement(job.siteMap.StartURL, link) {
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
						resultText := selectorElementAttribute(doc, &selector)
						linkOutput[selector.ID] = resultText
					} else if selector.Type == "SelectorImage" {
						resultText := selectorImage(doc, &selector)
						if len(resultText) != 0 {
							if len(resultText) == 1 {
								linkOutput[selector.ID] = resultText[0]
							} else {
								linkOutput[selector.ID] = resultText
							}
						}
					} else if selector.Type == "SelectorElement" {
						resultText := selectorElement(doc, &selector)
						linkOutput[selector.ID] = resultText
					} else if selector.Type == "SelectorTable" {
						resultText := selectorTable(doc, &selector)
						linkOutput[selector.ID] = resultText
					}
				}
			}
			job.linkOutput = linkOutput
			results <- job
		}
	}
}

func scraper(siteMap *scraping, parent string) map[string]interface{} {
	output := make(map[string]interface{})
	var wg sync.WaitGroup
	jobs := make(chan workerJob, settings.Workers)
	results := make(chan workerJob, settings.Workers)
	outputChannel := make(chan map[string]interface{})
	for x := 1; x <= settings.Workers; x++ {
		wg.Add(1)
		go worker(jobs, results, &wg)
	}
	go func() {
		fc := getURL(siteMap.StartURL)
		if fc != nil {
			for startURL := range fc {
				if !validURL(startURL) {
					continue
				}
				workerJob := workerJob{
					parent:   parent,
					startURL: startURL,
					siteMap:  siteMap,
				}
				jobs <- workerJob
			}
			close(jobs)
		}
	}()
	go func() {
		pageOutput := make(map[string]interface{})
		for job := range results {
			if len(job.linkOutput) != 0 {
				if job.parent == "_root" {
					out, err := ioutil.ReadFile(settings.OutputFile)
					if err != nil {
						logErrors(err)
						os.Exit(1)
					}
					var data = map[string]interface{}{}
					err = json.Unmarshal(out, &data)
					data[job.startURL] = job.linkOutput
					switch settings.Export {
					case "xml":
						output, err := xml.MarshalIndent(websiteData(job.linkOutput), "", "  ")
						if err != nil {
							logErrors(err)
							os.Exit(1)
						}
						_ = ioutil.WriteFile(settings.OutputFile, output, 0644)
					case "csv":
						csvFile, err := os.OpenFile(settings.OutputFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
						if err != nil {
							logErrors(err)
							os.Exit(1)
						}
						csvWriter := csv.NewWriter(csvFile)
						var rows [][]string
						for i, v := range data[job.startURL].(map[string]interface{}) {
							rows = append(rows, []string{i, fmt.Sprint(v)})
						}
						for _, row := range rows {
							err = csvWriter.Write(row)
							if err != nil {
								frontendLog(err)
								break
							}
						}
						csvWriter.Flush()
						err = csvFile.Close()

						if err != nil {
							frontendLog(err)
						}
					case "json":
						output, err := json.MarshalIndent(data, "", " ")
						if err != nil {
							logErrors(err)
							os.Exit(1)
						}
						_ = ioutil.WriteFile(settings.OutputFile, output, 0644)
					default:
						fmt.Println("Error: Please choose an output format.")
					}
				} else {
					pageOutput[job.startURL] = job.linkOutput
				}
			}
		}
		outputChannel <- pageOutput
	}()
	wg.Wait()
	close(results)
	output = <-outputChannel
	return output
}

func validURL(uri string) bool {
	_, err := url.ParseRequestURI(uri)
	return err == nil
}

func outputResult() {
	userFormat := strings.ToLower(settings.Export)
	allowedFormat := map[string]bool{
		"csv":  true,
		"xml":  true,
		"json": true,
	}
	if allowedFormat[userFormat] {
		err := ioutil.WriteFile(settings.OutputFile, []byte{}, 0644)
		if err != nil {
			logErrors(err)
		}
	} else {
		_, _ = fmt.Fprintf(os.Stderr, "Format \"%s\" not supported", userFormat)
		os.Exit(1)
	}
}

func scrape() {
	readJSON()
	clearCache()
	siteMap := sitemap
	outputResult()
	_ = scraper(&siteMap, "_root")
}
