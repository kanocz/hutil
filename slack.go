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
	slackimg   string
	slackemoji string
)

// SlackInit function sets slack team, token and channel for logging
func SlackInit(url, name, icon string) {
	slackurl = url
	slackname = name
	if strings.HasPrefix(icon, ":") {
		slackemoji = icon
	} else {
		slackurl = icon
	}
}

// SlackLog sends log message to slack's #log channel
func SlackLog(message string) error {
	if "" == slackurl {
		return errors.New("Slack url is not set")
	}
	go func() {
		data := map[string]string{"text": message}
		if "" != slackname {
			data["username"] = slackname
		}
		if "" != slackemoji {
			data["icon_emoji"] = slackemoji
		}
		if "" != slackurl {
			data["icon_url"] = slackurl
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
		resp.Body.Close()
	}()
	return nil
}
