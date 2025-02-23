package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
)

const (
	apiURL    = "https://api.notion.com/v1/pages"
	notionVer = "2022-06-28"
)

type NotionPage struct {
	Parent     Parent     `json:"parent"`
	Properties Properties `json:"properties"`
}

type Parent struct {
	DatabaseID string `json:"database_id"`
}

type Properties struct {
	Title Title `json:"Title"`
}

type Title struct {
	Title []Text `json:"title"`
}

type Text struct {
	Text Content `json:"text"`
}

type Content struct {
	Content string `json:"content"`
}

type app struct {
	fileName   string
	token      string
	databaseID string
}

func main() {
	fileName := flag.String("file-name", "", "file name")
	token := flag.String("token", "", "token for notion API")
	databaseID := flag.String("database-id", "", "target notion database ID")
	flag.Parse()

	a := &app{
		fileName:   *fileName,
		token:      *token,
		databaseID: *databaseID,
	}

	if err := a.run(); err != nil {
		log.Printf("%v\n", err)
		os.Exit(1)
	}
}

func (a *app) run() error {
	file, err := os.Open(a.fileName)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		row := scanner.Text()
		if err := a.createPage(row); err != nil {
			return fmt.Errorf("failed to create page %s: %w", row, err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to scan: %w", err)
	}
	return nil
}

func (a *app) createPage(title string) error {
	data := NotionPage{
		Parent: Parent{DatabaseID: a.databaseID},
		Properties: Properties{
			Title: Title{
				Title: []Text{
					{Text: Content{Content: title}},
				},
			},
		},
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to generate JSON: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to generate request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+a.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Notion-Version", notionVer)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("%s: response Status: %s\n", title, resp.Status)
	return nil
}
