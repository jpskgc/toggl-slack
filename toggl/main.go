package main

import (
	//"bytes"
	"context"
	//"encoding/json"
	"time"
	"net/http"
	"log"
	//"io/ioutil"
	"os"

	//"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)


func Handler(ctx context.Context) (*http.Response, error) {

	return GetTogglReports(), nil
}

func GetTogglReports() *http.Response{
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
	q.Add("until", yesterday)
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
    if err != nil {
        log.Fatal(err)
    }
	defer resp.Body.Close()

	return resp

    // data, err := ioutil.ReadAll(resp.Body)
    // if err != nil {
    //     log.Fatal(err)
	// }

	// return string(data)
}

func main() {
	lambda.Start(Handler)
}
