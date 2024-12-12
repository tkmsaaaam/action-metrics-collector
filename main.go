package main

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/slack-go/slack"
)

func main() {
	token := os.Getenv("SLACK_TOKEN")
	if token == "" {
		log.Println("SLACK_TOKEN is null")
		return
	}
	channelName := os.Getenv("CHANNEL_NAME")
	channelId := os.Getenv("CHANNEL_ID")
	if channelName == "" && channelId == "" {
		log.Println("CHANNEL_NAME and CHANNEL_ID are null")
		return
	}
	client := slack.New(token)
	channels, _, err := client.GetConversationsForUser(&slack.GetConversationsForUserParameters{})
	if err != nil {
		log.Println("can not get channel list", err)
		return
	}
	var c *slack.Channel = nil
	for _, channel := range channels {
		if channel.Name == channelName || channel.ID == channelId {
			c = &channel
			break
		}
	}

	if c == nil {
		log.Println("can not find channel")
		return
	}

	now := time.Now()
	oldest := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

	messagesResponse, err := client.GetConversationHistory(&slack.GetConversationHistoryParameters{ChannelID: c.ID, Limit: 1000, Oldest: strconv.FormatInt(oldest.Unix(), 10), Latest: strconv.FormatInt(now.Unix(), 10)})
	if err != nil {
		log.Println("can not get messages response", err)
		return
	}
	m := map[string]int{}
	for _, message := range messagesResponse.Messages {
		v, ok := m[message.Text]
		if ok {
			v = v + 1
		} else {
			m[message.Text] = 1
		}
	}
	log.Println("Result:")
	for k, v := range m {
		log.Println(k, ":", v)
	}
}
