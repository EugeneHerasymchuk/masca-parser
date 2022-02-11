package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/andybrewer/mack"
)

func main() {
		mascaURL := "https://api.volcanoteide.com/products/1927/availability/2021-12"
	mascaHeaders := []Header{
		{key: "x-api-key", value: "caminomasca"},
		{key: "accept", value: "application/json"},
	}

	client := &http.Client{}

	tick := time.Tick(30 * time.Second)
	for range tick {

		res, err := process(client, RequestInput{headers: mascaHeaders, url: mascaURL})

		if err != nil {
			fmt.Println("error", err)
		}

		if res {
			notifyPositive()
		}

		fmt.Println("not positive... ", time.Now())
	}

}

func notifyPositive() {
	mack.Say("Check maska tickets!")
	mack.Alert("Check masca!")
}

func process(client *http.Client, params RequestInput) (bool, error) {
	boom, err := callGET(client, params)

	if err != nil {
		return false, err
	}

	result, err := doWeHaveSlots(boom)

	if err != nil {
		return false, err
	}

	fmt.Println(result)

	return true, nil
}

type Header struct {
	key   string
	value string
}

type RequestInput struct {
	url     string
	headers []Header
}

type Data struct {
	Availability []struct {
		Date     string `json:"date"`
		Sessions []struct {
			Session     string `json:"session"`
			Consumed    int    `json:"consumed"`
			Available   int    `json:"available"`
			MaxQuantity int    `json:"max_quantity"`
			Highlighted bool   `json:"highlighted"`
		} `json:"sessions"`
	} `json:"availability"`
}

func callGET(client *http.Client, params RequestInput) (*Data, error) {
	req, err := http.NewRequestWithContext(context.Background(),
		http.MethodGet, params.url, nil)
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(params.headers); i++ {
		req.Header.Add(params.headers[i].key, params.headers[i].value)
	}

	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		panic(fmt.Sprintf("unexpected status: got %v", res.Status))
	}

	var boom Data

	err = json.NewDecoder(res.Body).Decode(&boom)
	if err != nil {
		panic(err)
	}

	return &boom, nil
}

func doWeHaveSlots(result *Data) (*time.Time, error) {
	const (
		layoutISO = "2006-01-02"
	)
	startDate, _ := time.Parse(layoutISO, "2021-12-06")
	endDate, _ := time.Parse(layoutISO, "2021-12-15")
	for _, day := range result.Availability {

		for _, session := range day.Sessions {

			t, _ := time.Parse(layoutISO, day.Date)
			if t.After(startDate) && t.Before(endDate) && session.Available > 0 {
				fmt.Println(t, session.Available)
				return &t, nil
			}
		}
	}
	return nil, errors.New("have not found")
}
