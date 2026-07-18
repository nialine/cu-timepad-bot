package timepad

import (
	"context"
	"encoding/json"
	"net/http"
	"regexp"

	"github.com/gocolly/colly/v2"
)

type Client struct {
	HTTPClient *http.Client
}

func (client *Client) GetRawData(ctx context.Context, url string) (string, error) {
	re := regexp.MustCompile("return *(?P<json>{.*});")

	c := colly.NewCollector(
		colly.StdlibContext(ctx),
	)

	c.SetClient(client.HTTPClient)

	var match []string

	c.OnHTML(".event-registration #evModel", func(e *colly.HTMLElement) {
		match = re.FindStringSubmatch(e.Text)
	})

	err := c.Visit(url)
	if err != nil {
		return "", err
	}

	if len(match) > 0 {
		jsonIndex := re.SubexpIndex("json")

		return match[jsonIndex], nil
	}

	return "", ErrParse
}

func (client *Client) GetData(ctx context.Context, url string) (*Event, error) {
	raw_json, err := client.GetRawData(ctx, url)
	if err != nil {
		return nil, err
	}

	var event Event
	err = json.Unmarshal([]byte(raw_json), &event)

	return &event, err
}
