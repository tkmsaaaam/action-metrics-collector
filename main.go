package main

import (
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
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
	m := map[string][]time.Time{}
	for _, message := range messagesResponse.Messages {
		if re := regexp.MustCompile(`[a-z0-9_]+`); !re.MatchString(message.Text) {
			continue
		}
		unixTime, err := strconv.ParseInt(strings.Split(message.Timestamp, ".")[0], 10, 64)
		if err != nil {
			continue
		}
		unixNanoTime, err := strconv.ParseInt(strings.Split(message.Timestamp, ".")[1], 10, 64)
		if err != nil {
			continue
		}
		t := time.Unix(unixTime, unixNanoTime)
		v, ok := m[message.Text]
		if ok {
			m[message.Text] = append(v, t)
		} else {
			m[message.Text] = []time.Time{t}
		}
	}
	log.Println("Result:")
	for k, v := range m {
		sort.Slice(v, func(i, j int) bool { return v[i].Unix() < v[j].Unix() })
		log.Println(k, ":", len((v)))
		var sum float64 = 0
		for i := 0; i < len(v); i++ {
			var d time.Duration = 0
			if i >= 1 {
				d = v[i].Sub(v[i-1])
				sum += d.Seconds()
			}
			log.Println(i+1, ":", v[i], ":", d)
		}
		if len(v) > 1 {
			avg := time.Duration(sum / float64(len(v)-1) * float64(time.Second))
			log.Println("average:", avg)
		}
	}
}
