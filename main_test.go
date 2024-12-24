package main

import (
	"sort"
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
