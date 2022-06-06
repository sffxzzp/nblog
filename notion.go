package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type (
	NotionAPI struct {
		BaseURL  string
		Secret   string
		Version  string
		Database string
	}
	PostData struct {
		ID      string    `json:"id"`
		Title   string    `json:"title"`
		Time    time.Time `json:"time"`
		Content string    `json:"content"`
		Plain   string    `json:"plain"`
	}
)

func newNotionAPI(secret string) *NotionAPI {
	return &NotionAPI{
		BaseURL: "https://api.notion.com",
		Secret:  secret,
		Version: "2021-08-16",
	}
}

func (n *NotionAPI) database(database string) {
	if len(database) < 36 {
		database = fmt.Sprintf("%s-%s-%s-%s-%s", database[:8], database[8:12], database[12:16], database[16:20], database[20:])
	}
	n.Database = database
}

func (n *NotionAPI) addHeaders(req *http.Request) {
	req.Header.Add("Authorization", "Bearer "+n.Secret)
	req.Header.Add("Notion-Version", n.Version)
	req.Header.Add("Content-Type", "application/json")
}

func (n *NotionAPI) getAllData() []PostData {
	start := ""
	var posts []PostData
	for {
		start, tposts := n.getData(start, 100)
		posts = append(posts, tposts...)
		if start == "" {
			break
		}
	}
	return posts
}

func (n *NotionAPI) getData(start string, pagesize int) (string, []PostData) {
	if start != "" {
		start = `"start_cursor": "` + start + `",`
	}
	url := n.BaseURL + "/v1/databases/" + n.Database + "/query"
	req, err := http.NewRequest("POST", url, bytes.NewBufferString(`{"sorts": [{ "property": "Time", "timestamp": "last_edited_time", "direction": "descending" }], `+start+` "page_size": `+strconv.Itoa(pagesize)+`}`))
	if err != nil {
		panic(err)
	}
	n.addHeaders(req)
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		panic(err)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	var recv map[string]interface{}
	json.Unmarshal(body, &recv)
	next, _ := recv["next_cursor"].(string)
	var posts []PostData
	for _, data := range recv["results"].([]interface{}) {
		tid := data.(map[string]interface{})["id"].(string)
		data, _ := data.(map[string]interface{})["properties"].(map[string]interface{})
		ts, err := time.Parse(time.RFC3339Nano, data["Time"].(map[string]interface{})["date"].(map[string]interface{})["start"].(string))
		if err != nil {
			panic(err)
		}
		published := data["Published"].(map[string]interface{})["checkbox"].(bool)
		rtf := data["Content"].(map[string]interface{})["rich_text"].([]interface{})
		var tHTML, outHTML, plain string
		for _, txt := range rtf {
			out := txt.(map[string]interface{})["plain_text"].(string)
			plain = plain + out
			href := fmt.Sprintf("%v", txt.(map[string]interface{})["href"])
			if href != "<nil>" {
				out = "<a href=\"" + href + "\">" + out + "</a>"
			}
			prop := txt.(map[string]interface{})["annotations"]
			if prop.(map[string]interface{})["bold"].(bool) {
				out = "<b>" + out + "</b>"
			}
			if prop.(map[string]interface{})["italic"].(bool) {
				out = "<i>" + out + "</i>"
			}
			if prop.(map[string]interface{})["strikethrough"].(bool) {
				out = "<del>" + out + "</del>"
			}
			if prop.(map[string]interface{})["underline"].(bool) {
				out = "<u>" + out + "</u>"
			}
			if prop.(map[string]interface{})["code"].(bool) {
				out = "<code>" + out + "</code>"
			}
			color := prop.(map[string]interface{})["color"].(string)
			if color == "default" {
				out = "<span>" + out + "</span>"
			} else {
				out = "<span style=\"color: " + color + " important;\">" + out + "</span>"
			}
			tHTML += out
		}
		tCont := strings.Split(tHTML, "\n")
		for _, t := range tCont {
			outHTML += "<p>" + t + "</p>"
		}
		if published {
			posts = append(posts, PostData{
				ID:      tid,
				Title:   data["Title"].(map[string]interface{})["title"].([]interface{})[0].(map[string]interface{})["plain_text"].(string),
				Time:    ts,
				Content: outHTML,
				Plain:   plain,
			})
		}
	}
	return next, posts
}
