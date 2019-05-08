package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
    "os"

	"github.com/aws/aws-lambda-go/lambda"
)

type Request struct {
	Records []struct {
		SNS struct {
			Type       string `json:"Type"`
			Timestamp  string `json:"Timestamp"`
			SNSMessage string `json:"Message"`
		} `json:"Sns"`
	} `json:"Records"`
}

type SNSMessage struct {
	Region              string `json:"region"`
	AccountID           string `json:"accountId"`
	EventTriggerName    string `json:"eventTriggerName"`
	ApplicationName     string `json:"applicationName"`
	DeploymentID        string `json:"deploymentId"`
	DeploymentGroupName string `json:"deploymentGroupName"`
	CreateTime          string `json:"createTime"`
	CompleteTime        string `json:"completeTime"`
	// DeploymentOverview  struct {
	// 	Failed     string `json:"Failed"`
	// 	InProgress string `json:"InProgress"`
	// 	Pending    string `json:"Pending"`
	// 	Skipped    string `json:"Skipped"`
	// 	Succeeded  string `json:"Succeeded"`
	// } `json:"deploymentOverview"`
	Status string `json:"status"`
}

type SlackMessage struct {
	Text        string       `json:"text"`
	Attachments []Attachment `json:"attachments"`
}

type Attachment struct {
	Text  string `json:"text"`
	Color string `json:"color"`
	Title string `json:"title"`
}

func handler(request Request) error {
	var snsMessage SNSMessage
	err := json.Unmarshal([]byte(request.Records[0].SNS.SNSMessage), &snsMessage)
	if err != nil {
		return err
	}

	log.Printf("Message: %s", snsMessage)
	slackMessage := buildSlackMessage(snsMessage)
	postToSlack(slackMessage)
	log.Println("Notification has been sent")
	return nil
}

func buildSlackMessage(message SNSMessage) SlackMessage {
    var color string
    SlackText := fmt.Sprintf("Deployment ID: %v Status: %v", message.DeploymentID, message.Status)
	if message.Status == "FAILED" || message.Status == "ABORTED" {
		color = "danger"
	} else {
		color = "good"
	}
	return SlackMessage{
		Text: fmt.Sprintf("`%s`", message.ApplicationName),
		Attachments: []Attachment{
			Attachment{
				Text:  SlackText,
				Color: color,
				Title: "Deployment Status",
			},
		},
	}
}

func postToSlack(message SlackMessage) error {
	client := &http.Client{}
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", os.Getenv("SLACK_WEBHOOK"), bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Println(resp.StatusCode)
		return err
	}

	return nil
}

func main() {
	lambda.Start(handler)
}
