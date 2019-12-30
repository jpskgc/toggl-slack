package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type Response events.APIGatewayProxyResponse

type togglData struct {
	TotalGrand      int `json:"total_grand"`
	TotalBillable   int `json:"total_billable"`
	TotalCurrencies []struct {
		Currency string `json:"currency"`
		Amount   int    `json:"amount"`
	} `json:"total_currencies"`
	Data []struct {
		ID    int `json:"id"`
		Title struct {
			Project string      `json:"project"`
			Client  interface{} `json:"client"`
		} `json:"title"`
		Time            int `json:"time"`
		TotalCurrencies []struct {
			Currency string `json:"currency"`
			Amount   int    `json:"amount"`
		} `json:"total_currencies"`
		Items []struct {
			Title struct {
				TimeEntry string `json:"time_entry"`
			} `json:"title"`
			Time int    `json:"time"`
			Cur  string `json:"cur"`
			Sum  int    `json:"sum"`
			Rate int    `json:"rate"`
		} `json:"items"`
	} `json:"data"`
}

type Slack struct {
	Text      string `json:"text"`
	Username  string `json:"username"`
	IconEmoji string `json:"icon_emoji"`
	Channel   string `json:"channel"`
}

func Handler(ctx context.Context) (Response, error) {
	var buf bytes.Buffer

	body := GetTogglReports()
	text := makeSurveyText(body)
	byte := sendToSlack(text)

	json.HTMLEscape(&buf, byte)

	resp := Response{
		StatusCode:      200,
		IsBase64Encoded: false,
		Body:            buf.String(),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}

	return resp, nil

}

func GetTogglReports() []byte {
	time.Sleep(1 * time.Second)

	client := &http.Client{Timeout: time.Duration(10) * time.Second}

	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	req, err := http.NewRequest("GET", "https://toggl.com/reports/api/v2/summary", nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("content-type", "application/json")
	req.SetBasicAuth(os.Getenv("toggl_api_token"), "api_token")

	q := req.URL.Query()
	q.Add("user_agent", os.Getenv("toggl_user_agent"))
	q.Add("workspace_id", os.Getenv("toggl_workspace_id"))
	q.Add("since", yesterday)
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	return data
}

func makeSurveyText(data []byte) string {
	var togglData togglData

	json.Unmarshal(data, &togglData)

	var text string

	totalGrand := float64(togglData.TotalGrand)
	text += "【合計時間: " + strconv.FormatFloat(totalGrand/1000/60/60, 'f', 1, 64) + "h】\n"

	for i, project := range togglData.Data {
		projectTime := float64(project.Time)

		text += ">*" + project.Title.Project + ": " + strconv.FormatFloat(projectTime/1000/60/60, 'f', 1, 64) + "h*\n"

		for _, item := range project.Items {
			time := float64(item.Time)
			text += ">・" + item.Title.TimeEntry + " [" + strconv.FormatFloat(time/1000/60/60, 'f', 1, 64) + "h]" + "\n"
		}

		if i != len(togglData.Data)-1 {
			text += "\n"
		}
	}

	return text
}

func sendToSlack(text string) []byte {
	params := Slack{
		Text:      text,
		Username:  "昨日やったこと",
		IconEmoji: ":dog_akitainu:",
		Channel:   os.Getenv("slack_channel_name"),
	}

	jsonparams, _ := json.Marshal(params)

	resp, err := http.PostForm(
		os.Getenv("slack_url"),
		url.Values{"payload": {string(jsonparams)}},
	)

	if err != nil {
		log.Fatal(err)
	}
	body, _ := ioutil.ReadAll(resp.Body)

	defer resp.Body.Close()

	return body
}

func main() {
	lambda.Start(Handler)
}
