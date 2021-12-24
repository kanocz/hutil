package hutil

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
)

var (
	slackurl   string
	slackname  string
	slackemoji string
	slackchan  string
)

// SlackInit function sets slack team, token and channel for logging
func SlackInit(url, name, icon, channel string) {
	slackurl = url
	slackname = name
	if strings.HasPrefix(icon, ":") {
		slackemoji = icon
	} else {
		slackurl = icon
	}
	slackchan = channel
}

// SlackLog sends log message to slack's #log channel
func SlackLog(message string) error {
	if slackurl == "" {
		return errors.New("slack url is not set")
	}
	go func() {
		data := map[string]string{"text": message}
		if slackname != "" {
			data["username"] = slackname
		}
		if slackemoji != "" {
			data["icon_emoji"] = slackemoji
		}
		if slackurl != "" {
			data["icon_url"] = slackurl
		}
		if slackchan != "" {
			data["channel"] = slackchan
		}

		jsondata, err := json.Marshal(data)

		if nil != err {
			log.Println("Error encoding json for slack message:", err)
			return
		}

		req, err := http.NewRequest("POST", slackurl, bytes.NewBuffer(jsondata))
		if nil != err {
			log.Println("Error while http.NewRequest for slack message:", err)
			return
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Println("Error sending message to slack:", err)
		}
		if nil != resp && nil != resp.Body {
			resp.Body.Close()
		}

	}()
	return nil
}
