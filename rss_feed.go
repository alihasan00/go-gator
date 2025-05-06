package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
)

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "gator")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var xmlData RSSFeed
	err = xml.Unmarshal(data, &xmlData)
	if err != nil {
		return nil, err
	}

	// Unescape HTML entities in channel data
	xmlData.Channel.Title = html.UnescapeString(xmlData.Channel.Title)
	xmlData.Channel.Description = html.UnescapeString(xmlData.Channel.Description)

	// Unescape HTML entities in each item
	for i := range xmlData.Channel.Item {
		xmlData.Channel.Item[i].Title = html.UnescapeString(xmlData.Channel.Item[i].Title)
		xmlData.Channel.Item[i].Description = html.UnescapeString(xmlData.Channel.Item[i].Description)
	}

	fmt.Println(xmlData)
	return &xmlData, nil
}
