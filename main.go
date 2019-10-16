package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type info struct {
	price string
	miles string
	year  int
}

func getOnePage(manufacturer string, model string, year int, page int) []info {

	log.Printf("Fetching page %d", page)

	client := &http.Client{}

	url := fmt.Sprintf("https://www.lacentrale.fr/listing?makesModelsCommercialNames=%s:%s&yearMax=%d&yearMin=%d&page=%d", manufacturer, model, year, year, page)
	// Request the HTML page.
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/76.0.3809.132 Safari/537.36 OPR/63.0.3368.107")

	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	replacer := strings.NewReplacer("&nbsp;", "", "km", "", "€", "", "\u00A0", "")

	result := make([]info, 0, 15)

	// Find the review items
	doc.Find(".kmYearPrice").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the band and title
		miles := replacer.Replace(s.Find(".fieldMileage").Text())
		price := replacer.Replace(s.Find(".fieldPrice").Text())

		result = append(result,
			info{miles: miles,
				price: price,
				year:  year})
	})

	return result
}

func getOneYear(manuf string, model string, year int, logToFile func([]info)) {
	log.Printf("Fetching year %d\n", year)
	for i := 1; i <= 6; i++ {
		data := getOnePage(manuf, model, year, i)
		logToFile(data)
		time.Sleep(2 * time.Second)
	}
}

func main() {
	manuf := flag.String("make", "BMW", "The make/manufacturer of the car.")
	model := flag.String("model", "X5", "The model of the car.")

	flag.Parse()

	year := 2009

	csvFile, err := os.OpenFile(fmt.Sprintf("%s_%s.csv", *manuf, *model),
		os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer csvFile.Close()

	fmt.Fprintln(csvFile, "Year, Price, Miles")

	logToFile := func(data []info) {
		for _, v := range data {
			fmt.Fprintf(csvFile, "%d, %s, %s\n", v.year, v.price, v.miles)
		}
	}

	for year < 2020 {
		getOneYear(*manuf, *model, year, logToFile)
		year++
	}
}
