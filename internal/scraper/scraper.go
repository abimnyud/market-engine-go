package scraper

import (
	"encoding/json"
	"fmt"
	"log"
	"market-engine-go/internal/infrastructure/repository"
	"market-engine-go/internal/models"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
)

func GetIdxTodaySnapshot() {
	shellPath := "./bin/chrome-headless-shell-mac-arm64/chrome-headless-shell"
	caps := selenium.Capabilities{"browserName": "chrome"}
	chromeCaps := chrome.Capabilities{
		Path: shellPath,
		Args: []string{
			"--headless",
			"--user-agent=Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36",
		},
	}

	caps.AddChrome(chromeCaps)

	path, err := exec.LookPath("chromedriver")
	if err != nil {
		log.Fatal("Could not find chromedriver")
	}

	service, err := selenium.NewChromeDriverService(path, 4444)
	if err != nil {
		log.Fatalf("Failed to start Selenium service: %v", err)
	}
	defer service.Stop()

	driver, err := selenium.NewRemote(caps, "")

	if err != nil {
		log.Fatal("Error:", err)
	}

	stockList := "https://www.idx.co.id/id/data-pasar/data-saham/daftar-saham"
	err = driver.Get(stockList)
	if err != nil {
		log.Fatal("Error:", err)
	}

	html, err := driver.PageSource()
	if err != nil {
		log.Fatal("Error:", err)
	}

	stocks := parseStocksData(html)

	todaysSnapshotUrl := "https://www.idx.co.id/id/data-pasar/ringkasan-perdagangan/ringkasan-saham"

	err = driver.Get(todaysSnapshotUrl)
	if err != nil {
		log.Fatal("Error:", err)
	}

	err = driver.WaitWithTimeout(func(wd selenium.WebDriver) (bool, error) {
		selector := "select[id^='vgt-select-rpp-']"
		dropdownTrigger, err := driver.FindElement(selenium.ByCSSSelector, selector)
		if err == nil {
			dropdownTrigger.Click()
			return true, nil
		}
		return false, nil

	}, 3*time.Second)

	if err != nil {
		log.Fatal("Could not find or click the dropdown button: ", err)
	}

	err = driver.WaitWithTimeout(func(wd selenium.WebDriver) (bool, error) {
		allOption, err := wd.FindElement(selenium.ByCSSSelector, "option[value='-1']")
		if err == nil {
			allOption.Click()
			return true, nil
		}
		return false, nil
	}, 3*time.Second)

	if err != nil {
		log.Fatal("Could not find or click the 'All' option:", err)
	}

	time.Sleep(3 * time.Second)

	snapshotHtml, err := driver.PageSource()
	if err != nil {
		log.Fatal("Error:", err)
	}

	snapshots, err := ParseTableToStocks(snapshotHtml)

	joined := joinStockLists(stocks, snapshots)
	repo := repository.NewCsvStockRepository("./output")
	repo.SaveAll(joined)
}

func ParseTableToStocks(html string) ([]models.Stock, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}

	var stocks []models.Stock

	// Find the table rows.
	// Usually tables on IDX have a 'tbody' and 'tr' tags.
	selection := doc.Find("#vgt-table")
	selection.Find("tbody tr").Each(func(i int, s *goquery.Selection) {
		// Extract each column (td)
		cols := s.Find("td")

		// 0: Kode Saham, 1: Tertinggi (High), 2: Terendah (Low),
		// 3: Penutupan (Close), 4: Selisih (Change), 5: Volume, 6: Nilai (Value), 7: Frekuensi
		if cols.Length() >= 8 {
			stock := models.Stock{
				Code:      strings.TrimSpace(cols.Eq(0).Text()),
				High:      strings.TrimSpace(cols.Eq(1).Text()),
				Low:       strings.TrimSpace(cols.Eq(2).Text()),
				Close:     strings.TrimSpace(cols.Eq(3).Text()),
				Change:    cleanNumber(strings.TrimSpace(cols.Eq(4).Text())),
				Volume:    strings.TrimSpace(cols.Eq(5).Text()),
				Value:     strings.TrimSpace(cols.Eq(6).Text()),
				Frequency: strings.TrimSpace(cols.Eq(7).Text()),
			}
			stocks = append(stocks, stock)
		}
	})

	return stocks, nil
}

func parseStocksData(html string) []models.Stock {
	reRowData := regexp.MustCompile(`(?s)rowData:(\[.*?\])`)
	match := reRowData.FindStringSubmatch(html)
	if len(match) < 2 {
		fmt.Println("No rowData found")
		return nil
	}
	rawRows := match[1]

	finalJson := aggressiveClean(rawRows)

	var stocks []models.Stock
	err := json.Unmarshal([]byte(finalJson), &stocks)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return nil
	}

	return stocks
}

func aggressiveClean(raw string) string {
	// Quote keys
	reKeys := regexp.MustCompile(`([{,])\s*(\w+):`)
	res := reKeys.ReplaceAllString(raw, `$1"$2":`)

	// Quote unquoted string values (skipping numbers and booleans)
	// Matches colon, optional space, then a word that doesn't start with " or digit
	reValues := regexp.MustCompile(`:\s*([^"{\[\d\s\-\.][^,}\s]*)\s*([,}\]])`)
	res = reValues.ReplaceAllString(res, `:"$1"$2`)

	res = regexp.MustCompile(`,\s*([}\]])`).ReplaceAllString(res, `$1`)

	return res
}

func joinStockLists(targets []models.Stock, sources []models.Stock) []models.Stock {
	sourceMap := make(map[string]models.Stock)
	for _, s := range sources {
		sourceMap[s.Code] = s
	}

	joinedList := make([]models.Stock, 0, len(targets))
	for _, t := range targets {
		if data, found := sourceMap[t.Code]; found {
			t.High = data.High
			t.Low = data.Low
			t.Close = data.Close
			t.Change = data.Change
			t.Volume = data.Volume
			t.Value = data.Value
			t.Frequency = data.Frequency
		}
		joinedList = append(joinedList, t)
	}

	return joinedList
}

func cleanNumber(val string) string {
	res := strings.ReplaceAll(val, ".", "")
	res = strings.ReplaceAll(res, "=", "")
	return strings.TrimSpace(res)
}
