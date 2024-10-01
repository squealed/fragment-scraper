package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func main() {
	baseurl := "https://fragment.com/username/"
	fmt.Println("Hello World")

	data, err := os.ReadFile("wordlist.txt")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}

	names := strings.Split(string(data), "\n")
	for i, name := range names {
		names[i] = strings.TrimSpace(name) // Remove both \r and \n, plus other whitespace characters
	}

	for _, name := range names {

		req, err := http.NewRequest("GET", baseurl+name, nil)
		if err != nil {
			fmt.Println("Error creating request for", name, ":", err)
			continue
		}

		referUrl := fmt.Sprintf("%s?query=%s", baseurl, name)

		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:126.0) Gecko/20100101 Firefox/126.0")
		req.Header.Set("X-Aj-Referer", referUrl)
		req.Header.Set("Referer", referUrl)
		req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
		req.Header.Set("Accept-Language", "en-US,en;q=0.5")
		req.Header.Set("X-Requested-With", "XMLHttpRequest")
		req.Header.Set("Priority", "u=1")
		req.Header.Set("DNT", "1")
		req.Header.Set("Connection", "keep-alive")
		req.Header.Set("Sec-Fetch-Dest", "empty")
		req.Header.Set("Sec-Fetch-Mode", "cors")
		req.Header.Set("Sec-Fetch-Site", "same-origin")
		req.Header.Set("TE", "trailers")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println("Error making request for", name, ":", err)
			continue
		}

		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response body for", name, ":", err)
			continue
		}

		var response map[string]interface{}
		if err := json.Unmarshal(body, &response); err != nil {
			fmt.Println("Error unmarshaling response for", name, ":", err)
			continue
		}

		hData, ok := response["h"].(string)
		if !ok {
			continue
		}

		if !strings.Contains(hData, "tm-section-header-status") {
			fmt.Println("Status section not found for", name)
			continue
		}

		status := extractStatus(hData)
		if status == "" {
			fmt.Println("Failed to extract status for", name)
			continue
		}

		if header, ok := response["h"].(string); ok {
			priceTON := extractTONPrice(header)
			priceUSD := extractUSDPrice(header)
			if priceTON <= 516 && status != "tm-status-taken" && status != "tm-status-unavail" {
				fmt.Printf("Username %s is under the threshold. TON: %f, USD: %f\n", name, priceTON, priceUSD)
			}
		}
	}
}

func extractStatus(hData string) string {
	parts := strings.Split(hData, "tm-section-header-status")
	if len(parts) < 2 {
		return ""
	}
	innerParts := strings.Split(parts[1], `">`)
	if len(innerParts) < 2 {
		return ""
	}
	return strings.TrimSpace(innerParts[0])
}

func extractTONPrice(header string) float64 {
	start := strings.Index(header, "icon-ton\">") + len("icon-ton\">")
	end := strings.Index(header[start:], "</div")
	if start == -1 || end == -1 {
		return 0
	}
	tonStr := header[start : start+end]

	tonStr = strings.ReplaceAll(tonStr, ",", "")

	tonPrice, _ := strconv.ParseFloat(tonStr, 64)
	return tonPrice
}

func extractUSDPrice(header string) float64 {
	start := strings.Index(header, "&#036;") + len("&#036;")
	end := strings.Index(header[start:], "</div")
	if start == -1 || end == -1 {
		return 0
	}
	usdStr := header[start : start+end]
	usdPrice, _ := strconv.ParseFloat(usdStr, 64)
	return usdPrice
}
