package luis

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// BASEURL is the base url for the luis api
const BASEURL = "https://api.projectoxford.ai/luis/v2.0/apps"

type Client struct {
	BaseURL         *url.URL
	client          *http.Client
	subscriptionKey string
}

type ParseResponse struct {
	Query             string             `json:"query"`
	TopScoringIntent  *Intent            `json:"topScoringIntent"`
	Intents           []*Intent          `json:"intents"`
	Entities          []*Entity          `json:"entities"`
	CompositeEntities []*CompositeEntity `json:"compositeEntities"`
}

type Intent struct {
	Intent string  `json:"intent"`
	Score  float64 `json:"score"`
}

type Entity struct {
	Entity     string  `json:"entity"`
	Type       string  `json:"type"`
	StartIndex int     `json:"startIndex"`
	EndIndex   int     `json:"endIndex"`
	Score      float64 `json:"score"`
}

type CompositeEntity struct {
	ParentType string                  `json:"parentType"`
	Value      string                  `json:"value"`
	Children   []*CompositeEntityChild `json:"children"`
}

type CompositeEntityChild struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

func NewClient(httpClient *http.Client, projectId, subscriptionKey string) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	baseURL, _ := url.Parse(fmt.Sprintf("%s/%s", BASEURL, projectId))

	return &Client{
		client:          httpClient,
		BaseURL:         baseURL,
		subscriptionKey: subscriptionKey,
	}
}

func (c *Client) Parse(query string) (*ParseResponse, error) {
	params := url.Values{
		"subscription-key": []string{c.subscriptionKey},
		"q":                []string{query},
	}
	c.BaseURL.RawQuery = params.Encode()

	var buf io.ReadWriter
	req, err := http.NewRequest("GET", c.BaseURL.String(), buf)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	parseResp := &ParseResponse{}
	err = json.NewDecoder(resp.Body).Decode(parseResp)
	if err == io.EOF {
		err = nil // ignore EOF errors caused by empty response body
	}

	return parseResp, err

}
