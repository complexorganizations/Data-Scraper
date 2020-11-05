### Data Scraper

Data Scraper is a super fast crawler, scraper used to scrape and extract data from anywhere. Its used to scrape data from wide ranges of applications.

---
### Installation

Lets first use `git` to download this repo
```
git clone https://github.com/complexorganizations/data-scraper.git
```
Than lets configure the scraper, open the `settings.json`
```
{"Log":true,"JavaScript":true,"Workers":10,"Captcha":[],"Proxy":[]}
```
After configuring the scraper you can copy your scraper rules to `sitemap.json`
```
{"_id":"prajwalkoirala.com","startUrl":["https://www.prajwalkoirala.com"],"selectors":[{"id":"name","type":"SelectorText","parentSelectors":["_root"],"selector":"h1","multiple":false,"regex":"","delay":0},{"id":"picture","type":"SelectorImage","parentSelectors":["_root"],"selector":"img","multiple":false,"delay":0}]}
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
- Dynamic Applications [JavaScript|ASP|AJAX]
- Proxy [HTTP|HTTPS]
- Docker
- Logging
- Captcha ***Coming Soon***
- Exports [JSON|XML|CSV] ***Coming Soon***

---
### Q&A

How do i use this?
- Download the [webscraper](https://webscraper.io/) extension, develop the scraper using the extension, export the scraper json rules after creating the scraper. ***Update: Developing your own solution to generate the sitemap.json***

How many domains can it scrape?
- This will scrape as many domains as you like. ***NO LIMITS***

How do i change what it scrapes?
- You can change what the scraper scrapes using `sitemap.json`

How do i configure the scraper?
- Open the settings file `settings.json` and change the scraper settings there.

Can this scrape apps written in JavaScript?
- Yes, this can scrape apps written in JS. ***Google Chrome (Required)***

Why not use a browser extension to scrape a website?
- The problem with browser extensions is that they are slow, and when it comes to large scraping projects it turns into a nightmare.

Why is this app so minimalist?
- Minimalist but extremely strong, this app is designed to run in the background.

---
### Concurrency
Most modern apps and system have rate limits.

| Workers         | Requests           | JavaScript         |
| --------------  | ------------------ | ------------------ |
| 1               | 300/min            | 75/min             |
| 5               | 1500/min           | 375/min            |
| 10              | 3000/min           | 750/min            |
| 25              | 7500/min           | 1875/min           |
| 50              | 15000/min          | 3750/min           |
| 100             | 30000/min          | 7500/min           |
| 500             | 150000/min         | 37500/min          |
| 1000            | 300000/min         | 75000/min          |

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
