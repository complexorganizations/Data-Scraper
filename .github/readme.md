### Data Scraper
Data Scraper is a super fast crawler, scraper used to scrape and extract data, Its used to scrape data from wide ranges of applications.

***Data Scraper is not yet complete. You should not rely on this code. It has not undergone proper degrees of security auditing and the protocol is still subject to change. We're working toward a stable 1.0.0 release, but that time has not yet come. There are experimental snapshots tagged with "0.0.0.MM-DD-YYYY", but these should not be considered real releases and they may contain security vulnerabilities (which would not be eligible for CVEs, since this is pre-release snapshot software). If you are packaging Data Scraper, you must keep up to date with the snapshots.***

---
### Installation

Lets first use `git` to download this repo
```
git clone https://github.com/complexorganizations/data-scraper.git
```
Here is a sample config, either build your own or use this one as an example. `sitemap.json`
```
{"settings":{"gui":true,"workers":1,"outputFile":"output.json"},"sitemap":{"projectID":"www.prajwalkoirala.com","startURL":["https://www.prajwalkoirala.com"],"selectors":[{"id":"Name","type":"SelectorText","parentSelectors":["_root"],"selector":"h1"},{"id":"Picture","type":"SelectorImage","parentSelectors":["_root"],"selector":"img"}]}}
```
You can finally run the scraper.
```
./data-scraper
```

---
### Features
- Unlimited [Crawling|Scraping|Parsing|Screenshot]
- Distributed
- Concurrency
- Dynamic Applications [JavaScript|ASP|AJAX|PHP]
- Proxy [HTTP|HTTPS]
- Docker
- Logging
- Captcha
- Exports [JSON|XML|CSV]

---
### Q&A

How do i use this?
- Use the GUI to build the `sitemap.json` and than use the scraper to start scraping.

How many domains can it scrape?
- This will scrape as many domains as you like.

How do i change what it scrapes?
- You can change what the scraper scrapes by generating new `sitemap.json`

How do i configure the scraper?
- Open the scraper and generate the `sitemap.json` file. 

Can this scrape apps written in JavaScript?
- Yes, this can scrape apps written in JS. ***Google Chrome (Required)***

Why not use a browser extension to scrape a website?
- The problem with browser extensions is that they are slow, and when it comes to large scraping projects it turns into a nightmare.

Why is this app so minimalist?
- Minimalist but extremely strong, this app is designed to run in the background.

Can this solve captcha?
- Yeah, captcha can be solved using Google cloud voice, vision API's.

Should i manually try and change the `sitemap.json` file?
- No, You should not.

How can i open the GUI again?
- Open the file `sitemap.json` and alter `"Gui": false` to `"Gui": true` ***Google Chrome (Required)***

---
### Concurrency
Most modern apps and system have rate limits.

| Workers         | Requests           | JavaScript         |
| --------------  | ------------------ | ------------------ |
| 1               | 250/min            | 50/min             |
| 5               | 1,250/min          | 250/min            |
| 10              | 2,500/min          | 500/min            |
| 25              | 6,250/min          | 1,250/min          |
| 50              | 12,500/min         | 2,500/min          |
| 100             | 25,000/min         | 5,000/min          |
| 500             | 125,000/min        | 25,000/min         |
| 1000            | 250,000/min        | 50,000/min         |

---
### Author

* Name: Prajwal Koirala
* Website: [prajwalkoirala.com](https://www.prajwalkoirala.com)
* Github: [@prajwal-koirala](https://github.com/prajwal-koirala)
* LinkedIn: [@prajwal-koirala](https://www.linkedin.com/in/prajwal-koirala)
* Twitter: [@Prajwal_K23](https://twitter.com/Prajwal_K23)
* Reddit: [@prajwalkoirala23](https://www.reddit.com/user/prajwalkoirala23)
* Twitch: [@prajwalkoirala23](https://www.twitch.tv/prajwalkoirala23)

---
### License

Copyright © 2020 [Prajwal](https://github.com/prajwal-koirala)

This project is MIT licensed.
