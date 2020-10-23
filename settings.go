package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/mohae/struct2csv"
	"gopkg.in/yaml.v2"
)

const (
	fileName = "settings"
)

type config struct {
	JavaScript    bool
	Proxy         bool
	ProxyLists    []string
	RotatingProxy bool
	Export        string
}

// Checks whether file exist or not
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

//save xml format
func (c *config) saveXMLFormat(data *config) {

	data.Export = "xml"

	outputFile := fmt.Sprintf("%s.%s", fileName, data.Export)

	if !fileExists(outputFile) {

		_, err := os.Create(outputFile)

		if err != nil {
			log.Fatal(err)
		}

	}

	xml, err := xml.MarshalIndent(data, "", " ")

	if err != nil {
		log.Fatal(err)
	}

	_ = ioutil.WriteFile(outputFile, xml, 0644)
}

//save JSON format
func (c *config) saveJSONFormat(data *config) {

	data.Export = "json"

	outputFile := fmt.Sprintf("%s.%s", fileName, data.Export)

	if !fileExists(outputFile) {

		_, err := os.Create(outputFile)

		if err != nil {
			log.Fatal(err)
		}

	}

	json, err := json.MarshalIndent(data, "", " ")

	if err != nil {
		log.Fatal(err)
	}

	_ = ioutil.WriteFile(outputFile, json, 0644)
}

//save txt format
func (c *config) saveTXTFormat(data *config) {

	data.Export = "txt"

	outputFile := fmt.Sprintf("%s.%s", fileName, data.Export)

	if !fileExists(outputFile) {

		_, err := os.Create(outputFile)

		if err != nil {
			log.Fatal(err)
		}

	}

	reqBodyBytes := new(bytes.Buffer)
	json.NewEncoder(reqBodyBytes).Encode(data)

	_ = ioutil.WriteFile(outputFile, reqBodyBytes.Bytes(), 0644)
}

// save CSV saveCsvFormat
func (c *config) saveCSVFormat(data *config) {

	data.Export = "csv"

	outputFile := fmt.Sprintf("%s.%s", fileName, data.Export)

	_, _ = os.Create(outputFile)

	content := []config{
		{
			JavaScript:    data.JavaScript,
			Proxy:         data.Proxy,
			ProxyLists:    data.ProxyLists,
			RotatingProxy: data.RotatingProxy,
			Export:        data.Export,
		},
	}
	encoder := struct2csv.New()
	csv, err := encoder.Marshal(content)
	if err != nil {
		fmt.Println("error:", err)
	}

	b := &bytes.Buffer{}
	w := struct2csv.NewWriter(b)
	w.WriteAll(csv)
	_ = ioutil.WriteFile(outputFile, b.Bytes(), 0644)

	w.Flush()
}

//save YAML format
//save JSON format
func (c *config) saveYAMLFormat(data *config) {

	data.Export = "yaml"

	outputFile := fmt.Sprintf("%s.%s", fileName, data.Export)

	if !fileExists(outputFile) {

		_, err := os.Create(outputFile)

		if err != nil {
			log.Fatal(err)
		}

	}

	yaml, err := yaml.Marshal(data)

	if err != nil {
		log.Fatal(err)
	}

	_ = ioutil.WriteFile(outputFile, yaml, 0644)
}

func main() {

	setConfig := &config{
		JavaScript:    false,
		Proxy:         false,
		ProxyLists:    []string{"socks5://127.0.0.1:8080", "http://localhost:8080"},
		RotatingProxy: false,
	}

	var format string

	fmt.Printf("Enter the format you wanted export(csv/json/txt/xml/yaml):")
	fmt.Scanf("%s", &format)

	switch strings.ToLower(format) {
	case "csv":
		setConfig.saveCSVFormat(setConfig)
	case "json":
		setConfig.saveJSONFormat(setConfig)
	case "txt":
		setConfig.saveTXTFormat(setConfig)
	case "xml":
		setConfig.saveXMLFormat(setConfig)
	case "yaml":
		setConfig.saveYAMLFormat(setConfig)
	}
}
