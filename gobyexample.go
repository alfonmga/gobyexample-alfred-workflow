package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/foolin/pagser"
)

type GobyexampleData struct {
	SectionsList []struct {
		Title string `pagser:"a"`
		Url   string `pagser:"a->attrConcat('href', 'https://gobyexample.com/', $value, '?from=gobyexample-alfred-workflow')"`
	} `pagser:"li"`
}

func fetchGobyexampleResults() (GobyexampleData, error) {
	resp, err := http.Get("https://gobyexample.com")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	p := pagser.New()

	var data GobyexampleData
	err = p.ParseReader(&data, resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	return data, nil
}

func (gobyexampleData GobyexampleData) toJSON() string {
	data, _ := json.MarshalIndent(gobyexampleData, "", "\t")
	return string(data)
}
func unmarshalGobyexampleDatafromJSON(blob []byte) GobyexampleData {
	data := GobyexampleData{}
	err := json.Unmarshal(blob, &data)
	if err != nil {
		wf.FatalError(err)
	}
	return data
}
