package main

import (
	"encoding/json"
	"fmt"
	"os"
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

func main() {
	os.Setenv("PAYLOAD", "{\"warehouse_locations\":[\"s3-hive-ireland\"],\"client_name\":\"tommy_hilfiger_pvh\",\"account_name\":\"tommy_hilfiger_gb_en\",\"report_name\":\"xml\",\"start_date\":\"2019-04-15\",\"provider_settings\":{\"url\":\"https://uk.tommy.com/sitemap_1_Home_en_GB.xml\",\"attributes_columns\":\"href,hreflang\",\"columns\":\"loc, xhtml:link, url\"}}")

	spayload := os.Getenv("PAYLOAD")
	var payload Payload
	json.Unmarshal([]byte(spayload), &payload)

	fmt.Println(payload)
	fmt.Println(payload.ProviderSettings)
}
