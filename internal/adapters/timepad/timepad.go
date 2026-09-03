package timepad

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strconv"

	"github.com/gocolly/colly/v2"
)

type Client struct {
	baseURL    *url.URL
	httpClient *http.Client
}

func New(baseurl *url.URL, httpclient *http.Client) *Client {
	return &Client{
		baseURL:    baseurl,
		httpClient: httpclient,
	}
}

func (client *Client) GetRawData(ctx context.Context, url string) (string, error) {
	re := regexp.MustCompile("return *(?P<json>{.*});")

	c := colly.NewCollector(
		colly.StdlibContext(ctx),
	)

	c.SetClient(client.httpClient)

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

func (client *Client) GetEventDataURL(ctx context.Context, url string) (*Event, error) {
	raw_json, err := client.GetRawData(ctx, url)
	if err != nil {
		return nil, err
	}

	var event Event
	err = json.Unmarshal([]byte(raw_json), &event)

	return &event, err
}

func (client *Client) GetEventDataID(ctx context.Context, eventid int64) (*Event, error) {
	if client.baseURL == nil {
		return nil, ErrNoBaseURL
	}

	url := client.baseURL.JoinPath("/event/" + strconv.FormatInt(eventid, 10))
	raw_json, err := client.GetRawData(ctx, url.String())
	if err != nil {
		return nil, err
	}

	var event Event
	err = json.Unmarshal([]byte(raw_json), &event)

	return &event, err
}

func (client *Client) GetCategoryID(ctx context.Context, magic_str string, slotid int64) (int64, error) {
	url := fmt.Sprintf("https://timepad.ru/api/event_model?callback=jQuery%s_0&response_type=jsonp&event=%d&_=0", magic_str, slotid)
	resp, err := client.httpClient.Get(url)
	if err != nil {
		return 0, ErrInvalid
	}
	body, _ := io.ReadAll(resp.Body)

	re := regexp.MustCompile("[[:alnum:]]*(?P<json>{.*})")
	match := re.FindStringSubmatch(string(body))
	if len(match) > 0 {
		var data map[string]any
		if err := json.Unmarshal([]byte(match[0]), &data); err != nil {
			return 0, ErrInvalid
		}

		if cat_id, ok := data["tickets"].([]any)[0].(map[string]any)["id"].(float64); ok {
			cat_id_int := int64(cat_id)
			slog.LogAttrs(ctx, slog.LevelDebug, "GetCategoryID", slog.Int64("cat_id", cat_id_int))
			return cat_id_int, nil
		}
	}

	return 0, ErrParse
}

func (client *Client) ReserveSlot(ctx context.Context, magic_str string, eventid, slotid int64, email, surname, name string) (*http.Response, error) {
	if client.baseURL == nil {
		return nil, ErrNoBaseURL
	}

	slog.LogAttrs(ctx, slog.LevelDebug, "ReserveSlot", slog.String("mail", email))
	post_url := client.baseURL.JoinPath(fmt.Sprintf("/event/widget_register/%v", slotid))

	category_id, err := client.GetCategoryID(ctx, magic_str, slotid)
	if err != nil {
		return nil, err
	}

	formData := url.Values{}
	formData.Set(fmt.Sprintf("res[%v]", category_id), "1")
	formData.Set("user_forms[0][mail]", email)
	formData.Set("user_forms[0][surname]", surname)
	formData.Set("user_forms[0][name]", name)
	formData.Set("payment_method", "yandex_sbp")
	formData.Set("accepted_terms", "on")
	formData.Set("tickets[0][re_id]", strconv.FormatInt(category_id, 10))
	formData.Set("locale", "ru")
	formData.Set("aux[use_ticker_remind]", "1")
	formData.Set("referer", "")

	slog.LogAttrs(ctx, slog.LevelDebug, "Timepad Reserve Slot",
		slog.String("FormData", formData.Encode()),
		slog.String("url", post_url.String()),
	)
	resp, err := client.httpClient.PostForm(post_url.String(), formData)
	data, err := io.ReadAll(resp.Body)
	slog.LogAttrs(ctx, slog.LevelDebug, "Timepad Reserve Slot", slog.String("ResponceBody", string(data)))
	if err != nil {
		return nil, err
	}

	return resp, err
}
