package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type prizeTier struct {
	ClaimedTickets     int    `json:"claimedTickets"`
	OriginalTierNumber int    `json:"originalTierNumber"`
	PaidTickets        int    `json:"paidTickets"`
	PrizeAmount        int    `json:"prizeAmount"`
	PrizeDescription   string `json:"prizeDescription"`
	TierNumber         int    `json:"tierNumber"`
	TierType           int    `json:"tierType"`
	WinningTickets     int    `json:"winningTickets"`
}

type game struct {
	DisableDate           uint64      `json:"disableDate"`
	EndDistributionDate   uint64      `json:"endDistributionDate"`
	GameID                string      `json:"gameId"`
	GameName              string      `json:"gameName"`
	PrizeTiers            []prizeTier `json:"prizeTiers"`
	StartDistributionDate uint64      `json:"startDistributionDate"`
	TicketPrice           int         `json:"ticketPrice"`
	TotalTicketsPrinted   int         `json:"totalTicketsPrinted"`
	ValidationStatus      string      `json:"validationStatus"`
}

func expectedValue(g game) float64 {
	var totalCents float64
	for _, prize := range g.PrizeTiers {
		totalCents += float64((prize.WinningTickets - prize.ClaimedTickets) * prize.PrizeAmount)
	}
	return totalCents/float64(100*g.TotalTicketsPrinted) - float64(g.TicketPrice)/100
}

type top struct {
	Games           []game        `json:"games"`
	NextItems       []interface{} `json:"nextItems"`
	NextPageURL     string        `json:"nextPageUrl"`
	PageURLs        []string      `json:"pageUrls"`
	PreviousItems   []interface{} `json:"previousItems"`
	PreviousPageURL string        `json:"previousPageUrl"`
}

func main() {
	req, err := http.NewRequest(http.MethodGet, "https://www.njlottery.com/api/v1/instant-games/games/?size=1000", nil)
	if err != nil {
		panic(err)
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/62.0.3202.94 Safari/537.36")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	js := &top{}
	json.NewDecoder(resp.Body).Decode(js)
	defer resp.Body.Close()

	var bestContestName string
	var bestContestTicketPrice float64
	var bestContestEV float64 = -100

	for _, g := range js.Games {
		if g.ValidationStatus == "ACTIVE" {
			ev := expectedValue(g)
			if bestContestEV < ev {
				bestContestName = g.GameName
				bestContestTicketPrice = float64(g.TicketPrice) / 100
				bestContestEV = ev
			}
		}
	}

	fmt.Printf("Contest: %s\nTicket Price ($): %0.2f\nExpected Value ($): %0.2f\n\n",
		bestContestName, bestContestTicketPrice, bestContestEV)
}
