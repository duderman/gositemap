package main

import (
	"bytes"
	"compress/gzip"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type ProviderSettings struct {
	URL string `json:"url"`
}

type Payload struct {
	ProviderSettings ProviderSettings `json:"provider_settings"`
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

const (
	S3_REGION    = "us-east-1"
	S3_BUCKET    = "sitemap-test"
	AWS_ENDPOINT = "http://127.0.0.1:9000/"
)

func awsEndpoint() string {
	envValue := os.Getenv("AWS_ENDPOINT")

	if len(envValue) > 0 {
		return envValue
	}

	return AWS_ENDPOINT
}

func (url URL) toTSV(sitemapURL string) [][]string {
	var link Link
	result := make([][]string, len(url.Links))

	for index := 0; index < len(url.Links); index++ {
		link = url.Links[index]
		result[index] = []string{url.Loc, link.Href, link.Lang, sitemapURL}
	}

	return result
}

func parsePayload() Payload {
	spayload := os.Getenv("PAYLOAD")
	var payload Payload
	json.Unmarshal([]byte(spayload), &payload)
	return payload
}

func newHTTPClient() *http.Client {
	tr := &http.Transport{DisableCompression: true}
	return &http.Client{Transport: tr}
}

func openSitemap(url string) (reader io.ReadCloser, err error) {
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

func parseXML(reader io.ReadCloser, out chan *URL) (err error) {
	defer close(out)

	fmt.Println("Parsing XML")

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

			out <- url
		case xml.EndElement:
			continue
		}
	}

	fmt.Println("Finished parsing")

	return nil
}

func openTSV(tsvFile *os.File) *csv.Writer {
	tsvOut := csv.NewWriter(tsvFile)
	tsvOut.Comma = '\t'

	return tsvOut
}

func writeToTSV(in chan *URL, finished chan string, sitemapURL string) {
	filename := "out.tsv"
	tsvFile, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer tsvFile.Close()

	tsv := openTSV(tsvFile)
	defer tsv.Flush()

	for url := range in {
		tsv.WriteAll(url.toTSV(sitemapURL))
	}

	fmt.Println("Finished output")
	finished <- filename
}

func newAWSConfig() *aws.Config {
	return &aws.Config{
		Region:           aws.String(S3_REGION),
		Endpoint:         aws.String(awsEndpoint()),
		DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(true),
	}
}

func uploadToS3(filename string) error {
	fmt.Println("Uploading to S3")
	config := newAWSConfig()
	sess, err := session.NewSession(config)

	if err != nil {
		return err
	}

	file, err := os.Open(filename)

	if err != nil {
		return err
	}
	defer file.Close()

	fileInfo, _ := file.Stat()
	size := fileInfo.Size()
	buffer := make([]byte, size)
	file.Read(buffer)

	fmt.Println("Collected all the info. Starting upload")
	_, err = s3.New(sess).PutObject(&s3.PutObjectInput{
		Bucket:             aws.String(S3_BUCKET),
		Key:                aws.String(filename),
		ACL:                aws.String("private"),
		Body:               bytes.NewReader(buffer),
		ContentLength:      aws.Int64(size),
		ContentType:        aws.String(http.DetectContentType(buffer)),
		ContentDisposition: aws.String("attachment"),
	})

	return err
}

func main() {
	fmt.Println("Starting")
	startTime := time.Now()
	os.Setenv("PAYLOAD", "{\"provider_settings\":{\"url\":\"https://www.ferragamo.com/sfsm/sitemap_33751.xml.gz\"}}")

	var err error

	fmt.Println("Parsing payload")
	payload := parsePayload()

	fmt.Println("Opening sitemap")
	url := payload.ProviderSettings.URL
	reader, err := openSitemap(url)
	defer reader.Close()

	if err != nil {
		panic(err)
	}

	urlsChan := make(chan *URL)
	finishedChan := make(chan string)

	go parseXML(reader, urlsChan)
	go writeToTSV(urlsChan, finishedChan, url)

	filename := <-finishedChan

	err = uploadToS3(filename)

	if err != nil {
		panic(err)
	}

	fmt.Println("Done")
	fmt.Println("It took: ", time.Since(startTime).Milliseconds())
}
