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

type Result struct {
	Details []*Detail
	sum     *time.Duration
}

type Detail struct {
	t    *time.Time
	diff *time.Duration
}

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
	m := makeMap(messagesResponse)
	print(m)
}

func makeMap(res *slack.GetConversationHistoryResponse) *map[string]*Result {
	m := map[string]*Result{}
	sort.Slice(res.Messages, func(i, j int) bool {
		return res.Messages[i].Timestamp < res.Messages[j].Timestamp
	})
	for _, message := range res.Messages {
		if re := regexp.MustCompile("^[a-z0-9_]+$"); !re.MatchString(message.Text) {
			continue
		}
		ary := strings.Split(message.Timestamp, ".")
		unixTime, err := strconv.ParseInt(ary[0], 10, 64)
		if err != nil {
			continue
		}
		unixNanoTime, err := strconv.ParseInt(ary[1], 10, 64)
		if err != nil {
			continue
		}
		t := time.Unix(unixTime, unixNanoTime)
		var d time.Duration = 0
		v, ok := m[message.Text]
		if ok {
			d = t.Sub(*(v.Details[len(v.Details)-1].t))
			sum := *(v.sum) + d
			m[message.Text] = &Result{Details: append(v.Details, &Detail{t: &t, diff: &d}), sum: &sum}
		} else {
			m[message.Text] = &Result{Details: []*Detail{{t: &t, diff: &d}}, sum: &d}
		}
	}
	return &m
}

func print(m *map[string]*Result) {
	log.Println("Result:")
	for k, v := range *m {
		log.Println(k, ":", len((v.Details)), "times")
		for i := 0; i < len(v.Details); i++ {
			log.Println(i+1, ":", v.Details[i].t, ":", v.Details[i].diff)
		}

		if len(v.Details) > 1 {
			log.Println("average:", (*v.sum).Seconds()/float64(len(v.Details)-1), "s")
		}
	}
}
