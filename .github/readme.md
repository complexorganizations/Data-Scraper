### Data Scraper

Data Scraper is a super fast crawler, scraper used to scrape and extract data from anywhere. Its used to scrape data from wide ranges of applications.

---
### Installation

Lets first use `git` to download this repo
```
git clone https://github.com/complexorganizations/Data-Scraper.git
```
Than lets configure the scraper, open the `settings.json`
```
JavaScript: true,false
Proxy: true,false
ProxyLists: ["socks5://127.0.0.1:8080","http://localhost:8080"]
RotatingProxy: true,false
Export: "json","csv","xml"
```
After configuring the scraper you can copy your scraper rules to `scraping.json`
```
{"_id":"prajwalkoirala.com","startUrl":["https://www.prajwalkoirala.com"],"selectors":[{"id":"name","type":"SelectorText","parentSelectors":["_root"],"selector":"h1","multiple":false,"regex":"","delay":0},{"id":"picture","type":"SelectorImage","parentSelectors":["_root"],"selector":"img","multiple":false,"delay":0}]}
```
You can finally run the scraper.
```
./Data-Scraper
```

---
### Features
- Unlimited scraping ***NO LIMITS***
- Distributed scraping
- Concurrency scraping
- JavaScript rendering ***Google Chrome (Required)***
- Dynamic applications
- Proxy support
- Docker support
- Exports to JSON

---
### Q&A

How do i use this?
- Download the [webscraper](https://webscraper.io/) extension, develop the scraper using the extension, export the scraper json rules after creating the scraper.

How fast is this?
- On our test, its about =>3k request a minute.

How many domains can it scrape?
- This will scrape as many domains as you like. ***NO LIMITS***

How do i change what it scrapes?
- You can change what the scraper scrapes using `scraping.json`

How do i configure the scraper?
- Open the settings file `settings.json` and change the scraper settings there.

Can this scrape apps written in JavaScript?
- Yes, this can scrape apps written in JS.

Why not use a browser extension to scrape a website?
- The problem with browser extensions is that they are slow, and when it comes to large scraping projects it turns into a nightmare.

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
