package main

import (
	"crypto/tls"
	"fmt"
	"github.com/zserge/lorca"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

var (
	shouldScrape = false
)

func frontendLog(err error) {
	if len(settings.LogFile) > 0 {
		_, _ = fmt.Fprintln(os.Stderr, "Error: ", err)
	}
}

func ifThenElse(condition bool, a string, b string) string {
	if condition {
		return a
	}
	return b
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

const globalStyles = `
	body {
		background-color: #f1f3f4;
		padding: 16px; 
	}
	th, td {
		padding: 3px;
	}
	table {
		width: 100%;
		background-color: #fff;
		border: 2px solid #dadce0;
		border-radius: 2px;
		padding: 1000px;
		border-collapse: collapse;
	}
	.buttons {
		padding: 10px;
	}
`

func editSitemap(ui lorca.UI) {
	err := ui.Load("data:text/html," + url.PathEscape(uiEditMap()))
	if err != nil {
		frontendLog(err)
	}
}

func viewSelectors(ui lorca.UI) {
	err := ui.Load("data:text/html," + url.PathEscape(uiViewSelectors()))
	if err != nil {
		frontendLog(err)
	}
}

func editSettings(ui lorca.UI) {
	err := ui.Load("data:text/html," + url.PathEscape(uiEditSettings()))
	if err != nil {
		frontendLog(err)
	}
}

func uiViewSitemap() string {
	page := `
		<html>
			<head>
				<title>Data Scraper Sitemap Generator</title>
				<style>
					` + globalStyles + `
					body {
						position: fixed;
						top: 0;
						bottom: 0;
						left: 0;
						right: 0;
						height: 100%;
						display: flex;
						flex-direction: column;
						align-items: center;
						justify-content: center;
					}
					table {
						width: 100%;
						max-width: 500px;
					}
				</style>
			</head>
			<body>
				<table>
					<tr>
						<th>ID</th>
						<th>Start URL</th>
					</tr>
					<tr>
						<td>` + sitemap.ID + `</td>
						<td>`
	for i, e := range sitemap.StartURL {
		page += e
		if i < len(sitemap.StartURL)-1 {
			page += ", "
		}
	}
	page += `</td>
					</tr>
				</table>
				<div class="buttons">
					<button onclick="editSitemap()">Edit Sitemap</button>
					<button onclick="viewSelectors()">View Selectors</button>
					<button onclick="editSettings()">Settings</button>
					<button onclick="runScraper()">Run</button>
				</div>
			</body>
		</html>
	`
	return page
}

func saveSettings(ui lorca.UI) {
	var err error
	settings.Gui = fmt.Sprint(ui.Eval(`document.getElementById("settings_gui").checked.toString();`)) == "true"
	settings.LogFile = fmt.Sprint(ui.Eval(`document.getElementById("settings_logfile").value;`))
	settings.JavaScript = newBool(fmt.Sprint(ui.Eval(`document.getElementById("settings_js").checked.toString();`)) == "true")
	settings.Workers, err = strconv.Atoi(fmt.Sprint(ui.Eval(`document.getElementById("settings_workers").value;`)))
	intA, err := strconv.Atoi(fmt.Sprint(ui.Eval(`document.getElementById("settings_rate_limit").value;`)))
	settings.RateLimit = newInt(intA)
	if err != nil {
		frontendLog(err)
	}
	uaNum, _ := strconv.Atoi(fmt.Sprint(ui.Eval(`user_agent_num.toString();`)))
	settings.UserAgents = []string{}
	for i := 0; i < uaNum; i++ {
		code := fmt.Sprintf(`document.getElementById("txt_useragent%d").value;`, i+1)
		settings.UserAgents = append(settings.UserAgents, fmt.Sprint(ui.Eval(code)))
	}
	settings.Captcha = fmt.Sprint(ui.Eval(`document.getElementById("settings_captcha").value;`))
	proxyNum, _ := strconv.Atoi(fmt.Sprint(ui.Eval(`proxy_num.toString();`)))
	settings.Proxy = []string{}
	for i := 0; i < proxyNum; i++ {
		code := fmt.Sprintf(`document.getElementById("txt_proxy%d").value;`, i+1)
		settings.Proxy = append(settings.Proxy, fmt.Sprint(ui.Eval(code)))
	}
	settings.OutputFile = fmt.Sprint(ui.Eval(`document.getElementById("settings_output").value;`))
	writeJSON()
	err = ui.Load("data:text/html," + url.PathEscape(uiViewSitemap()))
	if err != nil {
		frontendLog(err)
	}
}

func addUserAgent(ui lorca.UI) {
	ui.Eval(`
		(function() {
			if(user_agent_num > 0 && document.getElementById("txt_useragent" + user_agent_num).value.length == 0) {
				return;
			}
			user_agent_num++;
			el = document.createElement("input");
			el.id = "txt_useragent" + user_agent_num.toString();
			ua.appendChild(el);
		})()
	`)
}

func removeUserAgent(ui lorca.UI) {
	ui.Eval(`
		if(user_agent_num > 0) {
			user_agent_num--;
			ua.removeChild(ua.children[user_agent_num]);
		}
	`)
}

func addProxy(ui lorca.UI) {
	ui.Eval(`
		(function() {
			if(proxy_num > 0 && document.getElementById("txt_proxy" + proxy_num).value.length == 0) {
				return;
			}
			proxy_num++;
			el = document.createElement("input");
			el.id = "txt_proxy" + proxy_num.toString();
			proxies.appendChild(el);
		})()
	`)
}

func removeProxy(ui lorca.UI) {
	ui.Eval(`
		if(proxy_num > 0) {
			proxy_num--;
			proxies.removeChild(proxies.children[proxy_num]);
		}
	`)
}

func uiEditSettings() string {
	page := `
	<html>
		<head>
			<title>Edit settings</title>
			<style>
				` + globalStyles + `
				input {
					display: block;
				}
			</style>
		</head>
		<body>
			<table>
				<tr><th>Gui</th><td><input id="settings_gui" type="checkbox" ` + ifThenElse(settings.Gui, `checked`, "") + `></td></tr>
				<tr>
					<th>Log file</th>
					<td><input id="settings_logfile" type="text" value="` + settings.LogFile + `"></td>
				</tr>
				<tr><th>JavaScript</th><td><input id="settings_js" type="checkbox" ` + ifThenElse(*settings.JavaScript, `checked`, "") + `></td></tr>
				<tr><th>Workers</th><td><input id="settings_workers" type="number" value="` + strconv.Itoa(settings.Workers) + `"></td></tr>
				<tr><th>Rate limit</th><td><input id="settings_rate_limit" type="number" value="` + strconv.Itoa(*settings.RateLimit) + `"></td></tr>
				<tr><th>Output file</th><td><input id="settings_output" type="text" value="` + settings.OutputFile + `"></td></tr>
				<tr>
					<th>User agents</th>
					<td>
						<div id="userAgents">`
	for i, e := range settings.UserAgents {
		page += `<input type="text" id="txt_useragent` + strconv.Itoa(i+1) + `" value="` + e + `"></input>`
	}
	page += `	</div>
						<button onclick=removeUserAgent()>-</button>
						<button onclick=addUserAgent()>+</button>
					</td>
				</tr>
				<tr><th>Captcha</th><td><input id="settings_captcha" type="text" value="` + settings.Captcha + `"></td></tr>
				<tr>
					<th>Proxy</th>
					<td>
						<div id="proxies">`
	for i, e := range settings.Proxy {
		page += `<input type="text" id="txt_proxy` + strconv.Itoa(i+1) + `" value="` + e + `"></input>`
	}
	page += `	</div>
						<button onclick=removeProxy()>-</button>
						<button onclick=addProxy()>+</button>
					</td>
				</tr>
			</table>
			<div class="buttons">
				<button onclick="saveSettings()">Save</button>
			</div>
			<script>
				let user_agent_num = ` + strconv.Itoa(len(settings.UserAgents)) + `
				let ua = document.getElementById("userAgents");
				let proxy_num = ` + strconv.Itoa(len(settings.Proxy)) + `
				let proxies = document.getElementById("proxies");
				let el;
			</script>
		</body>
	</html>
	`
	return page
}

func addSiteURL(ui lorca.UI) {
	ui.Eval(`
		url_num++;
		el = document.createElement("input");
		el.id = "txt_starturl" + url_num.toString();
		url_inputs.appendChild(el);
	`)
}

func removeSiteURL(ui lorca.UI) {
	ui.Eval(`
		if(url_num > 1) {
			url_num--;
			url_inputs.removeChild(url_inputs.children[url_num]);
		}
	`)
}

func saveMap(ui lorca.UI) {
	sitemap.ID = fmt.Sprint(ui.Eval(`document.getElementById("txt_sitemap_id").value;`))
	urlNum, _ := strconv.Atoi(fmt.Sprint(ui.Eval(`url_num.toString();`)))
	sitemap.StartURL = []string{}
	for i := 0; i < urlNum; i++ {
		code := fmt.Sprintf(`document.getElementById("txt_starturl%d").value;`, i+1)
		sitemap.StartURL = append(sitemap.StartURL, fmt.Sprint(ui.Eval(code)))
	}

	if fmt.Sprint(ui.Eval(`document.getElementById("login").checked.toString();`)) == "true" {
		sitemap.Login = &login{
			URL:      fmt.Sprint(ui.Eval(`document.getElementById("txt_login_url").value;`)),
			Username: fmt.Sprint(ui.Eval(`document.getElementById("txt_login_username").value;`)),
			Password: fmt.Sprint(ui.Eval(`document.getElementById("txt_login_password").value;`)),
		}
	} else {
		sitemap.Login = nil
	}
	writeJSON()
	err := ui.Load("data:text/html," + url.PathEscape(uiViewSitemap()))
	if err != nil {
		frontendLog(err)
	}
}

func uiEditMap() string {
	page := `
		<html>
			<head>
				<title>Edit sitemap</title>
				<style>
					` + globalStyles + `
					input:not([type='checkbox']), label:not([for="login"]) {
						display: block;
					}
				.hide {
					display: none;
				}
				</style>
			</head>
			<body>
				<label for="txt_sitemap_id">Sitemap name: </label>
				<input type="text" placeholder="Enter sitemap name" id="txt_sitemap_id" value="` + sitemap.ID + `"></input>
				<label for="urlInputs">Start URL: </label>
				<div id="urlInputs">`
	for i, e := range sitemap.StartURL {
		page += `<input type="text" placeholder="Enter start URL" id="txt_starturl` + strconv.Itoa(i+1) + `" value="` + e + `"></input>`
	}
	sitemap.Login = &login{
		URL:      "",
		Username: "",
		Password: "",
	}
	page += `</div>
				<button onclick=removeSiteURL()>-</button>
				<button onclick=addSiteURL()>+</button>
				<br /><br />
				<label for="login">Require login</label>
				<input type="checkbox" id="login" ` + ifThenElse(sitemap.Login.URL == "", ``, `checked`) + `></input>
				<div id="show_login"  ` + ifThenElse(sitemap.Login.URL == "", ` class="hide"`, "") + `>
					<label for="txt_login_url">Login URL: </label>
					<input type="text" placeholder="Enter login url" id="txt_login_url" value="` + sitemap.Login.URL + `"></input>
					<label for="txt_login_username">Username: </label>
					<input type="text" placeholder="Enter username" id="txt_login_username" value="` + sitemap.Login.Username + `"></input>
					<label for="txt_login_password">Password: </label>
					<input type="text" placeholder="Enter password" id="txt_login_password" value="` + sitemap.Login.Password + `"></input>
				</div>
				<button onclick=saveMap()>Save</button>
				<script>
					let url_num = ` + strconv.Itoa(len(sitemap.StartURL)) + `
					let url_inputs = document.getElementById("urlInputs");
					let el;

					let checkbox = document.getElementById("login");
					let show_login = document.getElementById("show_login");
					checkbox.addEventListener('change', function() {
						show_login.classList.toggle("hide");
					});
				</script>
			</body>
		</html>
	`
	return page
}

func editSelector(ui lorca.UI, index int) {
	err := ui.Load("data:text/html," + url.PathEscape(uiEditSelector(index)))
	if err != nil {
		frontendLog(err)
	}
}

func viewMap(ui lorca.UI) {
	err := ui.Load("data:text/html," + url.PathEscape(uiViewSitemap()))
	if err != nil {
		frontendLog(err)
	}
}

func addSelector(ui lorca.UI) {
	newSelector := selectors{}
	newSelector.ParentSelectors = []string{""}
	sitemap.Selectors = append(sitemap.Selectors, newSelector)
	err := ui.Load("data:text/html," + url.PathEscape(uiEditSelector(len(sitemap.Selectors)-1)))
	if err != nil {
		frontendLog(err)
	}
}

func uiViewSelectors() string {
	page := `
		<html>
			<head>
				<title>View selectors</title>
				<style>
					` + globalStyles + `
				</style>
			</head>
			<body>
				<table>
					<tr>
						<th>id</th>
						<th>type</th>
						<th>parent selectors</th>
						<th>selector</th>
						<th>multiple</th>
						<th>regex</th>
						<th>delay</th>
						<th>edit</th>
					</tr>`
	for i, e := range sitemap.Selectors {
		page += `<tr>`
		page += `<td>` + e.ID + `</td>`
		page += `<td>` + e.Type + `</td>`
		page += `<td>`
		for i, el := range e.ParentSelectors {
			page += el
			if i < len(e.ParentSelectors)-1 {
				page += ", "
			}
		}
		page += `</td>`
		page += `<td>` + e.Selector + `</td>`
		if e.Multiple != nil && *e.Multiple {
			page += `<td> yes </td>`
		} else {
			page += `<td> no </td>`
		}
		page += `<td>` + e.Regex + `</td>`
		if e.Delay != nil {
			page += `<td>` + strconv.Itoa(*e.Delay) + `</td>`
		} else {
			page += `<td>0</td>`
		}

		page += `<td><button onclick="editSelector(` + strconv.Itoa(i) + `)">Edit</button></td>`
		page += `</tr>`
	}
	page += `</table>
				<div class="buttons">
					<button onclick=viewMap()>View sitemap</button>
					<button onclick=addSelector()>Add selector</button>
				</div>
			</body>
		</html>
	`
	return page
}

func deleteSelector(ui lorca.UI, index int) {
	sitemap.Selectors = append(sitemap.Selectors[:index], sitemap.Selectors[index+1:]...)
	writeJSON()
	err := ui.Load("data:text/html," + url.PathEscape(uiViewSelectors()))
	if err != nil {
		frontendLog(err)
	}
}

func saveSelector(ui lorca.UI, index int) {
	var err error
	el := sitemap.Selectors[index]
	el.ID = fmt.Sprint(ui.Eval(`document.getElementById("map_id").value;`))
	el.Type = fmt.Sprint(ui.Eval(`document.getElementById("map_type").value;`))
	el.ParentSelectors = []string{}
	parentNum, err := strconv.Atoi(fmt.Sprint(ui.Eval(`document.getElementById("map_parents").selectedOptions.length.toString();`)))
	for i := 0; i < parentNum; i++ {
		code := fmt.Sprintf(`document.getElementById("map_parents").selectedOptions[%d].value;`, i)
		el.ParentSelectors = append(el.ParentSelectors, fmt.Sprint(ui.Eval(code)))
	}
	el.Selector = fmt.Sprint(ui.Eval(`document.getElementById("map_selector").value;`))
	el.Multiple = newBool(fmt.Sprint(ui.Eval(`document.getElementById("map_multiple").checked.toString();`)) == "true")
	el.Regex = fmt.Sprint(ui.Eval(`document.getElementById("map_regex").value;`))
	intA, err := strconv.Atoi(fmt.Sprint(ui.Eval(`document.getElementById("map_delay").value;`)))
	el.Delay = newInt(intA)

	el.Download = newBool(fmt.Sprint(ui.Eval(`document.getElementById("download").checked.toString();`)) == "true")
	el.AttributeName = fmt.Sprint(ui.Eval(`document.getElementById("map_attr").value;`))
	el.HeaderRowSelector = fmt.Sprint(ui.Eval(`document.getElementById("map_hrs").value;`))
	el.DataRowsSelector = fmt.Sprint(ui.Eval(`document.getElementById("map_drs").value;`))
	el.FoundUrlRegex = fmt.Sprint(ui.Eval(`document.getElementById("map_fur").value;`))
	mp, err := strconv.ParseFloat(fmt.Sprint(ui.Eval(`document.getElementById("map_mip").value;`)), 64)
	el.MinimumPriority = newFloat64(mp)
	if err != nil {
		frontendLog(err)
	}
	sNum, _ := strconv.Atoi(fmt.Sprint(ui.Eval(`sitemap_num.toString();`)))
	el.SitemapURLs = []string{}
	for i := 0; i < sNum; i++ {
		code := fmt.Sprintf(`document.getElementById("txt_sitemap_url%d").value;`, i+1)
		el.SitemapURLs = append(el.SitemapURLs, fmt.Sprint(ui.Eval(code)))
	}
	el.ClickSelector = fmt.Sprint(ui.Eval(`document.getElementById("map_csl").value;`))
	el.ClickType = fmt.Sprint(ui.Eval(`document.getElementById("map_cty").value;`))
	el.ClickElementUnique = fmt.Sprint(ui.Eval(`document.getElementById("map_ceu").value;`))

	sitemap.Selectors[index] = el
	writeJSON()
	err = ui.Load("data:text/html," + url.PathEscape(uiViewSelectors()))
	if err != nil {
		frontendLog(err)
	}
}

func selectElement(ui lorca.UI, index int, selectURL string) {
	if selectURL == sitemap.StartURL[0] {
		saveSelector(ui, index)
	}
	err := ui.Load("data:text/html," + url.PathEscape(uiSelectElement(index, selectURL)))
	if err != nil {
		frontendLog(err)
	}
}

func addSitemap(ui lorca.UI) {
	ui.Eval(`
		(function() {
			if(sitemap_num > 0 && document.getElementById("txt_sitemap_url" + sitemap_num).value.length == 0) {
				return;
			}
			sitemap_num++;
			el = document.createElement("input");
			el.id = "txt_sitemap_url" + sitemap_num.toString();
			ua.appendChild(el);
		})()
	`)
}

func removeSitemap(ui lorca.UI) {
	ui.Eval(`
		if(sitemap_num > 0) {
			sitemap_num--;
			ua.removeChild(ua.children[sitemap_num]);
		}
	`)
}

func uiEditSelector(index int) string {
	el := sitemap.Selectors[index]
	page := `
		<html>
			<head>
				<title>Edit selectors</title>
				<style>
					` + globalStyles + `
					input{
						display: block;
					}
					.hide {
						display: none;
					}
				</style>
			</head>
			<body>
				<table>
					<tr><th>id</th><td><input type ="text" id="map_id" value="` + el.ID + `"></td></tr>
					<tr>
						<th>type</th><td>
						<select id="map_type">
							<option value="SelectorText" ` + ifThenElse(el.Type == "SelectorText", `selected`, "") + `>Selector Text</option>
							<option value="SelectorLink" ` + ifThenElse(el.Type == "SelectorLink", `selected`, "") + `>Selector Link</option>
							<option value="SelectorPopupLink" ` + ifThenElse(el.Type == "SelectorPopupLink", `selected`, "") + `>Selector Popup Link</option>
							<option value="SelectorImage" ` + ifThenElse(el.Type == "SelectorImage", `selected`, "") + `>Selector Image</option>
							<option value="SelectorTable" ` + ifThenElse(el.Type == "SelectorTable", `selected`, "") + `>Selector Table</option>
							<option value="SelectorElementAttribute" ` + ifThenElse(el.Type == "SelectorElementAttribute", `selected`, "") + `>Selector Element Attribute</option>
							<option value="SelectorHTML" ` + ifThenElse(el.Type == "SelectorHTML", `selected`, "") + `>Selector HTML</option>
							<option value="SelectorElement" ` + ifThenElse(el.Type == "SelectorElement", `selected`, "") + `>Selector Element</option>
							<option value="SelectorElementScroll" ` + ifThenElse(el.Type == "SelectorElementScroll", `selected`, "") + `>Selector Element Scroll</option>
							<option value="SelectorElementClick" ` + ifThenElse(el.Type == "SelectorElementClick", `selected`, "") + `>Selector Element Click</option>
							<option value="SelectorGroup" ` + ifThenElse(el.Type == "SelectorGroup", `selected`, "") + `>Selector Group</option>
							<option value="SelectorSitemapXmlLink" ` + ifThenElse(el.Type == "SelectorSitemapXmlLink", `selected`, "") + `>Selector Sitemap Xml Link</option>
						</select>
					</tr>
					<tr id="attr_tr"` + ifThenElse(el.Type == "SelectorElementAttribute", "", `class="hide"`) + `>
						<th>Attribute name</th>
						<td><input type ="text" id="map_attr" value="` + el.AttributeName + `"></td>
					</tr>
					<tr id="xml_tr"` + ifThenElse(el.Type == "SelectorSitemapXmlLink", "", `class="hide"`) + `>
						<th>Sitemap.xml Urls</th>
						<td>
							<div id="sitemaps">`
	for i, e := range el.SitemapURLs {
		page += `<input type="text" id="txt_sitemap_url` + strconv.Itoa(i+1) + `" value="` + e + `"></input>`
	}
	page += `</div>
							<button onclick=removeSitemap()>-</button>
							<button onclick=addSitemap()>+</button>
						</td>
					</tr>
					<tr id="fur_tr"` + ifThenElse(el.Type == "SelectorSitemapXmlLink", "", `class="hide"`) + `>
						<th>found url regex</th>
						<td><input type ="text" id="map_fur" value="` + el.FoundUrlRegex + `"></td>
					</tr>
					<tr id="mip_tr"` + ifThenElse(el.Type == "SelectorSitemapXmlLink", "", `class="hide"`) + `>
						<th>minimum priority</th>
						<td>`
	if el.MinimumPriority != nil {
		page += `<input type ="text" id="map_mip" value="` + fmt.Sprintf("%.2f", *el.MinimumPriority) + `">`
	} else {
		page += `<input type ="text" id="map_mip" value="">`
	}
	page += `</td>
					</tr>
					<tr id="csl_tr"` + ifThenElse(el.Type == "SelectorElementClick", "", `class="hide"`) + `>
						<th>Click selector</th>
						<td><input type ="text" id="map_csl" value="` + el.ClickSelector + `"></td>
					</tr>
					<tr id="cty_tr"` + ifThenElse(el.Type == "SelectorElementClick", "", `class="hide"`) + `>
						<th>Click type</th>
						<td>
							<select id="map_cty">
								<option value="once" ` + ifThenElse(el.ClickType == "once", `selected`, "") + `>Click once</option>
								<option value="more" ` + ifThenElse(el.ClickType == "more", `selected`, "") + `>Click more</option>
							</select>
						</td>
					</tr>
					<tr id="ceu_tr"` + ifThenElse(el.Type == "SelectorElementClick", "", `class="hide"`) + `>
						<th>Click element uniqueness</th>
						<td>
							<select id="map_ceu">
								<option value="text" ` + ifThenElse(el.Type == "text", `selected`, "") + `>text</option>
								<option value="htmlText" ` + ifThenElse(el.Type == "htmlText", `selected`, "") + `>html + text</option>
								<option value="html" ` + ifThenElse(el.Type == "html", `selected`, "") + `>html</option>
								<option value="css" ` + ifThenElse(el.Type == "css", `selected`, "") + `>css</option>
							</select>
						</td>
					</tr>
	`
	if el.Download != nil {
		page += `<tr id="download" ` + ifThenElse(el.Type == "SelectorImage", "", `class="hide"`) + `><th>Download</th><td><input type="checkbox" id="download"` + ifThenElse(*el.Download, "checked", "") + `></input></td></tr>`
	} else {
		page += `<tr id="download" ` + ifThenElse(el.Type == "SelectorImage", "", `class="hide"`) + `><th>Download</th><td><input type="checkbox" id="download"></input></td></tr>`
	}
	page += `<tr>
						<th>parent selectors</th>
						<td>
							<select id="map_parents" multiple>
								<option value="_root"` + ifThenElse(contains(el.ParentSelectors, "_root"), `selected="selected"`, "") + `>_root</option>`
	for _, e := range sitemap.Selectors {
		if e.ID != el.ID {
			page += `<option value="` + e.ID + `" ` + ifThenElse(contains(el.ParentSelectors, e.ID), `selected="selected"`, "") + `>` + e.ID + `</option>`
		}
	}
	page += `</select>
						</td>
					</tr>
					<tr>
						<th>selector</th>
						<td>
							<input type="text" id="map_selector" value="` + el.Selector + `">
							<button onclick="selectElement(` + strconv.Itoa(index) + `, '` + sitemap.StartURL[0] + `')">Select</button>
						</td>
					</tr>
					<tr id="hrs_tr"` + ifThenElse(el.Type == "SelectorTable", "", `class="hide"`) + `>
						<th>header row selectors</th>
						<td><input type ="text" id="map_hrs" value="` + el.HeaderRowSelector + `"></td>
					</tr>
					<tr id="drs_tr"` + ifThenElse(el.Type == "SelectorTable", "", `class="hide"`) + `>
						<th>data row selector</th>
						<td><input type ="text" id="map_drs" value="` + el.DataRowsSelector + `"></td>
					</tr>`
	if el.Multiple != nil {
		page += `<tr><th>multiple</th><td><input type="checkbox" id="map_multiple" ` + ifThenElse(*el.Multiple, `checked"`, "") + `></td></tr>`
	} else {
		page += `<tr><th>multiple</th><td><input type="checkbox" id="map_multiple"></td></tr>`
	}

	page += `<tr><th>regex</th><td><input type="text" id="map_regex" value="` + el.Regex + `"></td></tr>`
	if el.Delay != nil {
		page += `<tr><th>delay</th><td><input type="number" id="map_delay" value="` + strconv.Itoa(*el.Delay) + `"></td></tr>`
	} else {
		page += `<tr><th>delay</th><td><input type="number" id="map_delay" value="0"></td></tr>`
	}
	page += `</table>
				<div class="buttons">
					<button onclick=deleteSelector(` + strconv.Itoa(index) + `)>Delete</button>
					<button onclick=saveSelector(` + strconv.Itoa(index) + `)>Save</button>
				</div>
				<script>
					let sitemap_num = ` + strconv.Itoa(len(el.SitemapURLs)) + `
					let ua = document.getElementById("sitemaps");

					let select = document.getElementById("map_type");
					let download = document.getElementById("download");
					let attr_tr = document.getElementById("attr_tr");
					let hrs_tr = document.getElementById("hrs_tr");
					let drs_tr = document.getElementById("drs_tr");
					let xml_tr = document.getElementById("xml_tr");
					let fur_tr = document.getElementById("fur_tr");
					let mip_tr = document.getElementById("mip_tr");
					let csl_tr = document.getElementById("csl_tr");
					let cty_tr = document.getElementById("cty_tr");
					let ceu_tr = document.getElementById("ceu_tr");
					select.addEventListener('change', function() {
						download.classList.add("hide");
						attr_tr.classList.add("hide");
						hrs_tr.classList.add("hide");
						drs_tr.classList.add("hide");
						xml_tr.classList.add("hide");
						csl_tr.classList.add("hide");
						cty_tr.classList.add("hide");
						ceu_tr.classList.add("hide");
						switch(select.value) {
							case "SelectorImage":
								download.classList.remove("hide");		
								break;
							case "SelectorElementAttribute":
								attr_tr.classList.remove("hide");		
								break;
							case "SelectorTable":
								hrs_tr.classList.remove("hide");
								drs_tr.classList.remove("hide");		
								break;
							case "SelectorSitemapXmlLink":
								xml_tr.classList.remove("hide");
								fur_tr.classList.remove("hide");
								mip_tr.classList.remove("hide");		
								break;
							case "SelectorElementClick":
								csl_tr.classList.remove("hide");
								cty_tr.classList.remove("hide");
								ceu_tr.classList.remove("hide");		
								break;
						}
					});
				</script>
			</body>
		</html>
	`
	return page
}

func selectedElement(ui lorca.UI, index int, str string) {
	sitemap.Selectors[index].Selector = str
	editSelector(ui, index)
}

func uiSelectElement(index int, selectURL string) string {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
	}
	if len(settings.Proxy) > 0 {
		proxyString := settings.Proxy[0]
		proxyURL, _ := url.Parse(proxyString)
		transport.Proxy = http.ProxyURL(proxyURL)
	}
	client := &http.Client{Transport: transport}
	req, err := http.NewRequest("GET", selectURL, nil)
	if err != nil {
		frontendLog(err)
	}
	if len(settings.UserAgents) > 0 {
		req.Header.Set("User-Agent", settings.UserAgents[0])
	}
	resp, err := client.Do(req)
	var html []byte
	if err == nil {
		html, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			frontendLog(err)
		}
	}
	page := string(html)
	pURL, _ := url.Parse(selectURL)
	selectURL = pURL.Scheme + "://" + pURL.Host
	foundReplace := 0
	for true {
		var attrs = map[string]int{
			"href": -1,
			"src":  -1,
		}
		searchIndex := -1
		for attr := range attrs {
			attrs[attr] = strings.Index(page[foundReplace:], attr+"=\"")
			if attrs[attr] > -1 {
				attrs[attr] += len(page[:foundReplace])
			}
			if attrs[attr] == -1 {
				continue
			}
			if searchIndex == -1 {
				searchIndex = attrs[attr] + len(attr+"=\"")
			}
			if attrs[attr] < searchIndex {
				searchIndex = attrs[attr] + len(attr+"=\"")
			}
		}
		if searchIndex == -1 {
			break
		}
		if strings.HasPrefix(page[searchIndex:], "http") {
			foundReplace = searchIndex + 1
			continue
		}
		page = page[:searchIndex] + selectURL + ifThenElse(page[searchIndex] == '/', "", "/") + page[searchIndex:]
		foundReplace = searchIndex + 1
	}
	insertIndex := strings.Index(page, "</body>")
	if insertIndex == -1 {
		insertIndex = len(page) - 1
	}
	page =
		page[:insertIndex] +
			`<script defer>
			    let new_element;
			    let selected_elements_hover = [];
			    let selected_elements_new = [];
			    let identifiers;
			    let ui;
			    let choice_label;
			    let browse = true;
			    (function () {
			        document.querySelectorAll("a[href]").forEach((as) => {
			            as.onclick = (ev) => {
			                ev.preventDefault();
							if(as.href.includes("tel:") || as.href.includes("mailto:")) return;
			                selectElement(` + strconv.Itoa(index) + `, as.href);
			            }
			        })
			        ui = document.createElement("div");
			        choice_label = document.createElement("p");
			        choice_label.style.fontFamily = "sans-serif";
			        choice_label.style.flexGrow = "1";
			        ui.appendChild(choice_label);
			        let browse_button = document.createElement("button");
			        browse_button.style.fontFamily = "sans-serif";
			        browse_button.style.marginRight = "16px";
			        browse_button.onclick = () => {
			            document.onmouseover = browse ? mouseover : null;
			            browse = !browse;
			        }
			        browse_button.innerHTML = "Select element";
			        ui.appendChild(browse_button);
			        let accept_button = document.createElement("button");
			        accept_button.style.fontFamily = "sans-serif";
			        accept_button.onclick = () => selectedElement(` + strconv.Itoa(index) + `, identifiers);
			        accept_button.innerHTML = "Accept choice";
			        ui.appendChild(accept_button);
			        ui.style.position = "fixed";
			        ui.style.left = "0";
			        ui.style.right = "0";
			        ui.style.bottom = "0";
			        ui.style.zIndex = "10000";
			        ui.style.backgroundColor = "white";
			        ui.style.display = "flex";
			        ui.style.fontFamily = "sans-serif"
			        ui.style.padding = "8px 16px";
			        ui.style.borderTop = "solid black 1px";
			        document.body.appendChild(ui);
			    }())
			    const mouseover = (e) => {
			        const hoverIndex = el => selected_elements_hover.indexOf(el)
			        const newIndex = el => selected_elements_new.indexOf(el)
			        const removeSelected = i => {
			            selected_elements_new[i].remove()
			            selected_elements_hover.splice(i, 1);
			            selected_elements_new.splice(i, 1);
			        }
			        if (!!new_element && newIndex(new_element) == -1) new_element.remove();
			        let x = e.clientX, y = e.clientY;
			        let hover_element = document.elementFromPoint(x, y);
			        if (ui.contains(hover_element)) { hover_element = null; }
			        if (!!hover_element) {
			            new_element = document.createElement("div");
			            new_element.classList.add("test");
			            var rect = hover_element.getBoundingClientRect();
            			new_element.style.position = "absolute";
            			new_element.style.zIndex = "10000";
            			new_element.style.backgroundColor = "rgba(255, 0, 0, .2)";
            			new_element.style.top = rect.top + window.pageYOffset + "px";
			            new_element.style.height = rect.bottom - rect.top + "px";
			            new_element.style.left = rect.left + "px";
			            new_element.style.width = rect.right - rect.left + "px";
			            new_element.onmousedown = (e) => {
			                if (!hover_element) return;
							if (newIndex(hover_element) > -1) {
								removeSelected(newIndex(hover_element));
							} else if (hoverIndex(hover_element) > -1) {
								removeSelected(hoverIndex(hover_element));
							}
							else {
								selected_elements_new.push(new_element)
								selected_elements_hover.push(hover_element)
							}
			                identifiers = "";
			                selected_elements_hover.forEach((hover) => {
			                    if (identifiers.length > 0) {
			                        identifiers += ", ";
			                    }
			                    identifiers += hover.tagName.toLocaleLowerCase();
			                    if (hover.id.length > 0)
			                        identifiers += "#" + hover.id;
			
			                    hover.classList.forEach((e) => {
			                        identifiers += "." + e;
			                    })
			                })
			                choice_label.innerHTML = identifiers;
			                e.preventDefault();
			            }
			            document.body.appendChild(new_element);
			        }
			    }
			</script>` +
			page[insertIndex:]
	return page
}

func runScraper(ui lorca.UI) {
	shouldScrape = true
	err := ui.Close()
	if err != nil {
		frontendLog(err)
	}
}

func bindFunctions(ui lorca.UI) error {
	type binding struct {
		name     string
		function interface{}
	}
	functions := []binding{
		{"runScraper", func() { runScraper(ui) }},
		{"editSettings", func() { editSettings(ui) }},
		{"editSitemap", func() { editSitemap(ui) }},
		{"saveSettings", func() { saveSettings(ui) }},
		{"addUserAgent", func() { addUserAgent(ui) }},
		{"removeUserAgent", func() { removeUserAgent(ui) }},
		{"addProxy", func() { addProxy(ui) }},
		{"removeProxy", func() { removeProxy(ui) }},
		{"addSiteURL", func() { addSiteURL(ui) }},
		{"removeSiteURL", func() { removeSiteURL(ui) }},
		{"saveMap", func() { saveMap(ui) }},
		{"viewSelectors", func() { viewSelectors(ui) }},
		{"editSelector", func(i int) { editSelector(ui, i) }},
		{"deleteSelector", func(i int) { deleteSelector(ui, i) }},
		{"removeSitemap", func() { removeSitemap(ui) }},
		{"addSitemap", func() { addSitemap(ui) }},
		{"saveSelector", func(i int) { saveSelector(ui, i) }},
		{"addSelector", func() { addSelector(ui) }},
		{"viewMap", func() { viewMap(ui) }},
		{"selectElement", func(i int, url string) { selectElement(ui, i, url) }},
		{"selectedElement", func(i int, str string) { selectedElement(ui, i, str) }},
	}
	var err error = nil
	for _, e := range functions {
		err = ui.Bind(e.name, e.function)
		if err != nil {
			return err
		}
	}
	return nil
}

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
