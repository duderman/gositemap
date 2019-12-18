package main

import (
	"compress/gzip"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type ProviderSettings struct {
	AttributesColumns string `json:"attributes_columns"`
	Columns           string `json:"columns"`
	URL               string `json:"url"`
}

type Payload struct {
	WarehouseLocations []string         `json:"warehouse_locations"`
	ClientName         string           `json:"client_name"`
	AccountName        string           `json:"account_name"`
	ReporterName       string           `json:"reporter_name"`
	StartDate          string           `json:"start_date"`
	ProviderSettings   ProviderSettings `json:"provider_settings"`
}

func parsePayload() Payload {
	spayload := os.Getenv("PAYLOAD")
	var payload Payload
	json.Unmarshal([]byte(spayload), &payload)
	return payload
}

func tmpFilename() string {
	now := time.Now().Unix()

	return fmt.Sprintf("sitemap%d.xml", now)
}

func newHTTPClient() *http.Client {
	tr := &http.Transport{DisableCompression: true}
	return &http.Client{Transport: tr}
}

func downloadFile(url string) (filepath string, err error) {
	// Get the data
	client := newHTTPClient()
	resp, err := client.Get(url)
	if err != nil {
		return filepath, err
	}
	defer resp.Body.Close()

	// Support compression
	var reader io.ReadCloser
	switch resp.Header.Get("Content-Encoding") {
	case "x-gzip":
		reader, err = gzip.NewReader(resp.Body)
		defer reader.Close()
	default:
		reader = resp.Body
	}

	// Create the file
	filepath = tmpFilename()

	out, err := os.Create(filepath)
	if err != nil {
		return filepath, err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, reader)
	if err != nil {
		return filepath, err
	}

	return filepath, nil
}

func main() {
	os.Setenv("PAYLOAD", "{\"warehouse_locations\":[\"s3-hive-ireland\"],\"client_name\":\"tommy_hilfiger_pvh\",\"account_name\":\"tommy_hilfiger_gb_en\",\"report_name\":\"xml\",\"start_date\":\"2019-04-15\",\"provider_settings\":{\"url\":\"https://uk.tommy.com/sitemap_1_Home_en_GB.xml\",\"attributes_columns\":\"href,hreflang\",\"columns\":\"loc, xhtml:link, url\"}}")

	payload := parsePayload()

	url := payload.ProviderSettings.URL
	filepath, err := downloadFile(url)

	if err != nil {
		panic(err)
	}

	xmlFile, err := os.Open(filepath)

	if err != nil {
		panic(err)
	}

	xmlDec := xml.NewDecoder(xmlFile)

	for {
		t, tokenErr := xmlDec.Token()
		if tokenErr != nil {
			if tokenErr == io.EOF {
				break
			} else {
				panic(tokenErr.Error())
			}
		}
		switch startElem := t.(type) {
		case xml.StartElement:
			fmt.Println(startElem)
		case xml.EndElement:
			continue
		}
	}
}
