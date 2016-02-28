package hutil

import (
	"bytes"
	"log"
	"net/http"
)

var (
	slackurl string
)

// SlackInit function sets slack team, token and channel for logging
func SlackToken(team, token, channel string) {
	slackurl = "https://" + team + ".slack.com/services/hooks/slackbot?token=" + token + "&channel=%23" + channel
}

// SlackLog sends log message to slack's #log channel
func SlackLog(message string) {
	go func() {
		if "" == slackurl {
			log.Println("Slack token is not set yet")
			return
		}
		req, err := http.NewRequest("POST", slackurl, bytes.NewBuffer([]byte(message)))
		if nil == err {
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				log.Println("Error sending message to slack:", err)
			}
			resp.Body.Close()
		}
	}()
}
