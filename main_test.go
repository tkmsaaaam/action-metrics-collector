package main

import (
	"bytes"
	"log"
	"os"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/slack-go/slack"
)

func TestMakeMap(t *testing.T) {

	a := time.Date(2017, 12, 1, 8, 52, 30, 0, time.Local)
	b := 0 * time.Second
	c := time.Date(2017, 12, 1, 8, 52, 40, 0, time.Local)
	d := 10 * time.Second
	e := time.Date(2017, 12, 1, 8, 52, 50, 0, time.Local)
	f := time.Date(2017, 12, 1, 8, 53, 0, 0, time.Local)
	g := 30 * time.Second

	tests := []struct {
		name   string
		apiRes *slack.GetConversationHistoryResponse
		want   map[string]*Result
	}{
		{
			name:   "invalid event",
			apiRes: &slack.GetConversationHistoryResponse{Messages: []slack.Message{{Msg: slack.Msg{Text: "あいうえお", Timestamp: "1512085950.000000"}}}},
			want:   map[string]*Result{},
		},
		{
			name:   "invalid time",
			apiRes: &slack.GetConversationHistoryResponse{Messages: []slack.Message{{Msg: slack.Msg{Text: "test", Timestamp: "a.000000"}}}},
			want:   map[string]*Result{},
		},
		{
			name:   "invalid nano time",
			apiRes: &slack.GetConversationHistoryResponse{Messages: []slack.Message{{Msg: slack.Msg{Text: "test", Timestamp: "1512085950.a"}}}},
			want:   map[string]*Result{},
		},
		{
			name:   "one event",
			apiRes: &slack.GetConversationHistoryResponse{Messages: []slack.Message{{Msg: slack.Msg{Text: "test", Timestamp: "1512085950.000000"}}}},
			want:   map[string]*Result{"test": {Details: []*Detail{{t: &a, diff: &b}}, sum: &b}},
		},
		{
			name:   "two events",
			apiRes: &slack.GetConversationHistoryResponse{Messages: []slack.Message{{Msg: slack.Msg{Text: "test", Timestamp: "1512085950.000000"}}, {Msg: slack.Msg{Text: "test", Timestamp: "1512085960.000000"}}}},
			want:   map[string]*Result{"test": {Details: []*Detail{{t: &a, diff: &b}, {t: &c, diff: &d}}, sum: &d}},
		},
		{
			name:   "Multiple events of multiple types",
			apiRes: &slack.GetConversationHistoryResponse{Messages: []slack.Message{{Msg: slack.Msg{Text: "another", Timestamp: "1512085950.000000"}}, {Msg: slack.Msg{Text: "test", Timestamp: "1512085960.000000"}}, {Msg: slack.Msg{Text: "test", Timestamp: "1512085970.000000"}}, {Msg: slack.Msg{Text: "another", Timestamp: "1512085980.000000"}}}},
			want:   map[string]*Result{"test": {Details: []*Detail{{t: &c, diff: &b}, {t: &e, diff: &d}}, sum: &d}, "another": {Details: []*Detail{{t: &a, diff: &b}, {t: &f, diff: &g}}, sum: &g}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := makeMap(tt.apiRes)
			if got == nil {
				t.Errorf("makeMap() = %v, want %v", *got, tt.want)
			}
			if len(*got) != len(tt.want) {
				t.Errorf("makeMap() = %v, want %v", len(*got), len(tt.want))
			}
			for k, v := range tt.want {
				if _, ok := (*got)[k]; !ok {
					t.Errorf("makeMap() = %v, want %v", (*got)[k], tt.want)
				}
				if len((*got)[k].Details) != len(v.Details) {
					t.Errorf("makeMap() len(details) = %v, want %v", len((*got)[k].Details), len(v.Details))
				}
				if *((*got)[k].sum) != *(v.sum) {
					t.Errorf("makeMap() sum = %v, want %v", *((*got)[k].sum), *(v.sum))
				}
				sort.Slice(v.Details, func(i, j int) bool {
					return v.Details[i].t.Unix() < v.Details[j].t.Unix()
				})
				sort.Slice((*got)[k].Details, func(i, j int) bool {
					return (*got)[k].Details[i].t.Unix() < (*got)[k].Details[j].t.Unix()
				})
				for i := 0; i < len(v.Details); i++ {
					if v.Details[i].t.Unix() != (*got)[k].Details[i].t.Unix() {
						t.Errorf("makeMap() unix = %v, want %v", v.Details[i].t.Unix(), (*got)[k].Details[i].t.Unix())
					}
					if *(v.Details[i].diff) != *((*got)[k].Details[i].diff) {
						t.Errorf("makeMap() diff = %v, want %v", *(v.Details[i].diff), *((*got)[k].Details[i].diff))
					}
				}
			}
		})
	}
}

func TestPrint(t *testing.T) {
	a := 0 * time.Second
	b := time.Date(2017, 12, 1, 8, 52, 30, 0, time.Local)
	c := time.Date(2017, 12, 1, 8, 52, 40, 0, time.Local)
	d := 10 * time.Second
	e := time.Date(2017, 12, 1, 8, 52, 50, 0, time.Local)
	f := time.Date(2017, 12, 1, 8, 53, 0, 0, time.Local)

	tests := []struct {
		name string
		m    map[string]*Result
		want string
	}{
		{
			name: "empty",
			m:    map[string]*Result{},
			want: "Result:",
		},
		{
			name: "one event",
			m:    map[string]*Result{"test": {Details: []*Detail{{t: &b, diff: &a}}, sum: &a}},
			want: "Result:\ntest : 1 times\n1 : 2017-12-01 08:52:30 +0900 JST : 0s",
		},
		{
			name: "two events",
			m:    map[string]*Result{"test": {Details: []*Detail{{t: &b, diff: &a}, {t: &c, diff: &d}}, sum: &d}},
			want: "Result:\ntest : 2 times\n1 : 2017-12-01 08:52:30 +0900 JST : 0s\n2 : 2017-12-01 08:52:40 +0900 JST : 10s\naverage: 10 s",
		},
		{
			name: "Multiple events of multiple types",
			m:    map[string]*Result{"test": {Details: []*Detail{{t: &b, diff: &a}, {t: &c, diff: &d}}, sum: &d}, "another": {Details: []*Detail{{t: &e, diff: &a}, {t: &f, diff: &d}}, sum: &d}},
			want: "Result:\ntest : 2 times\n1 : 2017-12-01 08:52:30 +0900 JST : 0s\n2 : 2017-12-01 08:52:40 +0900 JST : 10s\naverage: 10 s\nanother : 2 times\n1 : 2017-12-01 08:52:50 +0900 JST : 0s\n2 : 2017-12-01 08:53:00 +0900 JST : 10s\naverage: 10 s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Helper()

			var buf bytes.Buffer
			log.SetOutput(&buf)
			defaultFlags := log.Flags()
			log.SetFlags(0)
			defer func() {
				log.SetOutput(os.Stderr)
				log.SetFlags(defaultFlags)
				buf.Reset()
			}()
			print(&tt.m)
			gotPrint := strings.TrimRight(buf.String(), "\n")
			if gotPrint != tt.want {
				t.Errorf("print() = \n%v,\nwant \n%v", gotPrint, tt.want)
			}
		})
	}
}
