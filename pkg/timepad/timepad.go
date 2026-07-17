package timepad

import (
	"context"
	"encoding/json"
	"regexp"

	"github.com/gocolly/colly/v2"
)

func GetRawData(ctx context.Context, url string) (string, error) {
	c := colly.NewCollector(
		colly.StdlibContext(ctx),
	)

	var raw_json string

	c.OnHTML("script #id", func(e *colly.HTMLElement) {
		re, err := regexp.Compile("return *(?P<json>{.*});")
		if err != nil {
			panic("ERROR: Compilation of regex failed")
		}
		raw_json = re.FindString(e.Text)
	})

	c.Visit(url)

	return raw_json, nil
}

func GetData(ctx context.Context, url string) (*Event, error) {
	raw_json, err := GetRawData(ctx, url)
	if err != nil {
		return nil, err
	}

	var event Event
	err = json.Unmarshal([]byte(raw_json), &event)

	return &event, err
}
