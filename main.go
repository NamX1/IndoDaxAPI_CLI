package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/valyala/fasthttp"
)

const baseEndpoint = "https://indodax.com"

type ServerTimeResponse struct {
	ServerTime int64  `json:"server_time"`
	Timezone   string `json:"timezone"`
}

type TickerResponse struct {
	Ticker struct {
		High       float64 `json:"high"`
		Low        float64 `json:"low"`
		Volume     float64 `json:"volume"`
		Last       float64 `json:"last"`
		Buy        float64 `json:"buy"`
		Sell       float64 `json:"sell"`
		ServerTime int64   `json:"server_time"`
	} `json:"ticker"`
}

type Pair struct {
	ID           string `json:"id"`
	Symbol       string `json:"symbol"`
	BaseCurrency string `json:"base_currency"`
	Description  string `json:"description"`
}

func processResponse(statusCode int, body []byte) ([]byte, error) {
	if statusCode == 200 {
		return body, nil
	}
	return nil, fmt.Errorf("unable to fetch data. status code: %d", statusCode)
}

func formatServerTime(serverTime int64) string {
	return time.Unix(serverTime/1000, 0).UTC().Format("2006-01-02 15:04:05")
}

func clearScreen() {
	os.Stdout.WriteString("\x1b[3;J\x1b[H\x1b[2J")
}

func main() {
	info := color.New(color.FgYellow).SprintFunc()
	success := color.New(color.FgGreen).SprintFunc()
	prompt := color.New(color.FgCyan).SprintFunc()
	errorMsg := color.New(color.FgRed).SprintFunc()
	header := color.New(color.FgBlue).Add(color.Underline).SprintFunc()

	fmt.Printf("[%s] This program is using %s!\n", info("i"), success("Official IndoDax API"))
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("Use %s if you don't know the commands.\n--> ", prompt("help"))
		user, _ := reader.ReadString('\n')
		user = strings.TrimSpace(user)

		switch user {
		case "help":
			fmt.Println(header("[i] Command List"))
			fmt.Printf("[%s] %s | Clear prompt\n", header("="), "cls")
			fmt.Printf("[%s] %s | Provide server time on exchange\n", header("="), "stime")
			fmt.Printf("[%s] %s | Provide available pairs on exchange\n", header("="), "pairs")

		case "clear":
			fmt.Printf("[%s] Clearing prompt.\n", success("+"))
			time.Sleep(1 * time.Second)
			clearScreen()

		case "servertime":
			statusCode, body, err := fasthttp.Get(nil, baseEndpoint+"/api/server_time")
			if err != nil {
				fmt.Printf("[%s] Unable to fetch data: %s\n", errorMsg("!"), err)
				continue
			}
			body, err = processResponse(statusCode, body)
			if err != nil {
				fmt.Printf("[%s] %s\n", errorMsg("!"), err)
				continue
			}
			var serverTimeData ServerTimeResponse
			json.Unmarshal(body, &serverTimeData)
			formattedTime := formatServerTime(serverTimeData.ServerTime)
			fmt.Printf("[%s] Server Time: %s %s\n", success("+"), formattedTime, serverTimeData.Timezone)

		case "pairs":
			statusCode, body, err := fasthttp.Get(nil, baseEndpoint+"/api/pairs")
			if err != nil {
				fmt.Printf("[%s] Unable to fetch data: %s\n", errorMsg("!"), err)
				continue
			}
			body, err = processResponse(statusCode, body)
			if err != nil {
				fmt.Printf("[%s] %s\n", errorMsg("!"), err)
				continue
			}
			var pairsData []Pair
			json.Unmarshal(body, &pairsData)
			for _, pair := range pairsData {
				fmt.Printf("[%s] ID: %s\n", success("+"), pair.ID)
				fmt.Printf("[%s] Symbol: %s\n", header("-"), pair.Symbol)
				fmt.Printf("[%s] Base Currency: %s\n", header("-"), pair.BaseCurrency)
				fmt.Printf("[%s] Description: %s\n\n", header("-"), pair.Description)
			}

		case "ticker":
			fmt.Print("Enter pair symbol (e.g., btcidr): ")
			pairSymbol, _ := reader.ReadString('\n')
			pairSymbol = strings.TrimSpace(pairSymbol)
			statusCode, body, err := fasthttp.Get(nil, baseEndpoint+"/api/ticker/"+pairSymbol)
			if err != nil {
				fmt.Printf("[%s] Unable to fetch data: %s\n", errorMsg("!"), err)
				continue
			}
			body, err = processResponse(statusCode, body)
			if err != nil {
				fmt.Printf("[%s] %s\n", errorMsg("!"), err)
				continue
			}
			var tickerData struct {
				Ticker struct {
					High       string `json:"high"`
					Low        string `json:"low"`
					VolTen     string `json:"vol_ten"`
					VolIDR     string `json:"vol_idr"`
					Last       string `json:"last"`
					Buy        string `json:"buy"`
					Sell       string `json:"sell"`
					ServerTime int64  `json:"server_time"`
				} `json:"ticker"`
			}
			err = json.Unmarshal(body, &tickerData)
			if err != nil {
				fmt.Printf("[%s] Error parsing ticker data: %s\n", errorMsg("!"), err)
				continue
			}
			fmt.Printf("[%s] Last Price: %s\n", success("+"), tickerData.Ticker.Last)
			fmt.Printf("[%s] High: %s\n", header("-"), tickerData.Ticker.High)
			fmt.Printf("[%s] Low: %s\n", header("-"), tickerData.Ticker.Low)
			fmt.Printf("[%s] Volume in TEN: %s\n", header("-"), tickerData.Ticker.VolTen)
			fmt.Printf("[%s] Volume in IDR: %s\n", header("-"), tickerData.Ticker.VolIDR)
			fmt.Printf("[%s] Buy: %s\n", header("-"), tickerData.Ticker.Buy)
			fmt.Printf("[%s] Sell: %s\n\n", header("-"), tickerData.Ticker.Sell)

		default:
			fmt.Printf("[%s] Unknown command\n", errorMsg("!"))
		}
	}
}
