package main

import (
	"compress/gzip"
	"encoding/csv"
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

type Link struct {
	XMLName xml.Name `xml:"link"`
	Lang    string   `xml:"hreflang,attr"`
	Href    string   `xml:"href,attr"`
}

type URL struct {
	Loc   string `xml:"loc"`
	Links []Link `xml:"link"`
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

func openSitemap(url string) (reader io.ReadCloser, err error) {
	// Get the data
	client := newHTTPClient()
	resp, err := client.Get(url)
	if err != nil {
		return reader, err
	}

	// Support compression
	switch resp.Header.Get("Content-Encoding") {
	case "x-gzip", "gzip":
		reader, err = gzip.NewReader(resp.Body)
	default:
		reader = resp.Body
	}

	return reader, err
}

func parseXML(reader io.ReadCloser) (err error) {
	xmlDec := xml.NewDecoder(reader)

	for {
		t, tokenErr := xmlDec.Token()
		if tokenErr != nil {
			if tokenErr == io.EOF {
				break
			} else {
				return tokenErr
			}
		}
		switch startElem := t.(type) {
		case xml.StartElement:
			if startElem.Name.Local != "url" {
				continue
			}

			url := &URL{}

			err = xmlDec.DecodeElement(url, &startElem)

			if err != nil {
				return err
			}

		case xml.EndElement:
			continue
		}
	}

	return nil
}

func main() {
	os.Setenv("PAYLOAD", "{\"warehouse_locations\":[\"s3-hive-ireland\"],\"client_name\":\"tommy_hilfiger_pvh\",\"account_name\":\"tommy_hilfiger_gb_en\",\"report_name\":\"xml\",\"start_date\":\"2019-04-15\",\"provider_settings\":{\"url\":\"https://uk.tommy.com/sitemap_1_Home_en_GB.xml\",\"attributes_columns\":\"href,hreflang\",\"columns\":\"loc, xhtml:link, url\"}}")

	var err error

	payload := parsePayload()

	url := payload.ProviderSettings.URL
	reader, err := openSitemap(url)
	defer reader.Close()

	if err != nil {
		panic(err)
	}

	err = parseXML(reader)

	if err != nil {
		panic(err)
	}

	tsvFile, err := os.Create("out.tsv")
	defer tsvFile.Close()

	if err != nil {
		panic(err)
	}

	tsvOut := csv.NewWriter(tsvFile)
	tsvOut.Comma = '\t'
	defer tsvOut.Flush()

	tsvOut.Write([]string{"zalupa", "konya"})
	tsvOut.Write([]string{"ad", "pizdy"})
	tsvOut.Write([]string{"1"})

	fmt.Println("Done")
}
