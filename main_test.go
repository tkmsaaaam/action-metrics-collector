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

	tests := []struct {
		name   string
		apiRes *slack.GetConversationHistoryResponse
		want   map[string]*Result
	}{
		{
			name:   "test",
			apiRes: &slack.GetConversationHistoryResponse{Messages: []slack.Message{{Msg: slack.Msg{Text: "test", Timestamp: "1512085950.000000"}}}},
			want:   map[string]*Result{"test": {Details: []*Detail{{t: &a, diff: &b}}, sum: &b}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := makeMap(tt.apiRes)
			if got == nil {
				t.Errorf("makeMap() = %v, want %v", *got, tt.want)
			}
			if len(*got) != len(tt.want) {
				t.Errorf("makeMap() = %v, want %v", *got, tt.want)
			}
			for k, v := range tt.want {
				if _, ok := (*got)[k]; !ok {
					t.Errorf("makeMap() = %v, want %v", *got, tt.want)
				}
				if len((*got)[k].Details) != len(v.Details) {
					t.Errorf("makeMap() = %v, want %v", *got, tt.want)
				}
				if *((*got)[k].sum) != *(v.sum) {
					t.Errorf("makeMap() = %v, want %v", *got, tt.want)
				}
				sort.Slice(v.Details, func(i, j int) bool {
					return v.Details[i].t.Unix() < v.Details[j].t.Unix()
				})
				sort.Slice((*got)[k].Details, func(i, j int) bool {
					return (*got)[k].Details[i].t.Unix() < (*got)[k].Details[j].t.Unix()
				})
				for i := 0; i < len(v.Details); i++ {
					if v.Details[i].t.Unix() != (*got)[k].Details[i].t.Unix() {
						t.Errorf("makeMap() = %v, want %v", *got, tt.want)
					}
					if *(v.Details[i].diff) != *((*got)[k].Details[i].diff) {
						t.Errorf("makeMap() = %v, want %v", *got, tt.want)
					}
				}
			}
		})
	}
}
