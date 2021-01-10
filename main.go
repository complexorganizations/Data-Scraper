package main

import (
	"data-scraper/cmd/backend"
	"data-scraper/cmd/frontend"
)

func main() {
	readJSON()
	if !settings.Gui {
		scrape()
		return
	}
	ui, err := lorca.New("", "", 900, 600)
	if err != nil {
		frontendLog(err)
		return
	}
	err = bindFunctions(ui)
	if err != nil {
		frontendLog(err)
	}
	err = ui.Load("data:text/html," + url.PathEscape(uiViewSitemap()))
	if err != nil {
		frontendLog(err)
	}
	<-ui.Done()
	err = ui.Close()
	if err != nil {
		frontendLog(err)
	}
	if shouldScrape {
		scrape()
	}
}
