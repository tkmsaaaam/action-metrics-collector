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

var (
	zero                           = 0 * time.Second
	ten                            = 10 * time.Second
	date_1512085950_20171201085230 = time.Date(2017, 12, 1, 8, 52, 30, 0, time.Local)
	date_1512085960_20171201085240 = time.Date(2017, 12, 1, 8, 52, 40, 0, time.Local)
	date_1512085970_20171201085250 = time.Date(2017, 12, 1, 8, 52, 50, 0, time.Local)
	date_1512085980_20171201085300 = time.Date(2017, 12, 1, 8, 53, 0, 0, time.Local)
)

func TestMakeMap(t *testing.T) {
	thirty := 30 * time.Second

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
			name:   "invalid time format",
			apiRes: &slack.GetConversationHistoryResponse{Messages: []slack.Message{{Msg: slack.Msg{Text: "test", Timestamp: ""}}}},
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
			name:   "valid event name",
			apiRes: &slack.GetConversationHistoryResponse{Messages: []slack.Message{{Msg: slack.Msg{Text: "test-example", Timestamp: "1512085950.000000"}}}},
			want:   map[string]*Result{"test-example": {Details: []*Detail{{t: &date_1512085950_20171201085230, diff: &zero}}, sum: &zero}},
		},
		{
			name:   "valid event name",
			apiRes: &slack.GetConversationHistoryResponse{Messages: []slack.Message{{Msg: slack.Msg{Text: "Test", Timestamp: "1512085950.000000"}}}},
			want:   map[string]*Result{"Test": {Details: []*Detail{{t: &date_1512085950_20171201085230, diff: &zero}}, sum: &zero}},
		},
		{
			name:   "valid event name",
			apiRes: &slack.GetConversationHistoryResponse{Messages: []slack.Message{{Msg: slack.Msg{Text: "test_example", Timestamp: "1512085950.000000"}}}},
			want:   map[string]*Result{"test_example": {Details: []*Detail{{t: &date_1512085950_20171201085230, diff: &zero}}, sum: &zero}},
		},
		{
			name:   "one event",
			apiRes: &slack.GetConversationHistoryResponse{Messages: []slack.Message{{Msg: slack.Msg{Text: "test", Timestamp: "1512085950.000000"}}}},
			want:   map[string]*Result{"test": {Details: []*Detail{{t: &date_1512085950_20171201085230, diff: &zero}}, sum: &zero}},
		},
		{
			name:   "two events",
			apiRes: &slack.GetConversationHistoryResponse{Messages: []slack.Message{{Msg: slack.Msg{Text: "test", Timestamp: "1512085950.000000"}}, {Msg: slack.Msg{Text: "test", Timestamp: "1512085960.000000"}}}},
			want:   map[string]*Result{"test": {Details: []*Detail{{t: &date_1512085950_20171201085230, diff: &zero}, {t: &date_1512085960_20171201085240, diff: &ten}}, sum: &ten}},
		},
		{
			name:   "Multiple events of multiple types",
			apiRes: &slack.GetConversationHistoryResponse{Messages: []slack.Message{{Msg: slack.Msg{Text: "another", Timestamp: "1512085950.000000"}}, {Msg: slack.Msg{Text: "test", Timestamp: "1512085960.000000"}}, {Msg: slack.Msg{Text: "test", Timestamp: "1512085970.000000"}}, {Msg: slack.Msg{Text: "another", Timestamp: "1512085980.000000"}}}},
			want:   map[string]*Result{"test": {Details: []*Detail{{t: &date_1512085960_20171201085240, diff: &zero}, {t: &date_1512085970_20171201085250, diff: &ten}}, sum: &ten}, "another": {Details: []*Detail{{t: &date_1512085950_20171201085230, diff: &zero}, {t: &date_1512085980_20171201085300, diff: &thirty}}, sum: &thirty}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := makeMap(tt.apiRes)
			actual := *got
			if len(actual) != len(tt.want) {
				t.Errorf("makeMap() = %v, want %v", len(actual), len(tt.want))
			}
			for k, v := range tt.want {
				result, ok := actual[k]
				if !ok {
					t.Errorf("%v is not found in %v", k, actual)
				}
				if len(result.Details) != len(v.Details) {
					t.Errorf("makeMap() len(actual) = %v, len(want) %v", len(result.Details), len(v.Details))
				}
				if *(result.sum) != *(v.sum) {
					t.Errorf("makeMap() sum = %v, want %v", *(result.sum), *(v.sum))
				}
				sort.Slice(v.Details, func(i, j int) bool {
					return v.Details[i].t.Unix() < v.Details[j].t.Unix()
				})
				sort.Slice(result.Details, func(i, j int) bool {
					return result.Details[i].t.Unix() < result.Details[j].t.Unix()
				})
				for i := 0; i < len(v.Details); i++ {
					v.Details[i].assert(result.Details[i], t)
				}
			}
		})
	}
}

func (d *Detail) assert(other *Detail, t *testing.T) {
	if d.t.Unix() != other.t.Unix() {
		t.Errorf("makeMap() unix = %v, want %v", d.t.Unix(), other.t.Unix())
	}
	if *(d.diff) != *(other.diff) {
		t.Errorf("makeMap() diff = %v, want %v", *(d.diff), *(other.diff))
	}
}

func TestPrint(t *testing.T) {
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
			m:    map[string]*Result{"test": {Details: []*Detail{{t: &date_1512085950_20171201085230, diff: &zero}}, sum: &zero}},
			want: "Result:test : 1 times1 : 2017-12-01 08:52:30 +0900 JST : 0s",
		},
		{
			name: "two events",
			m:    map[string]*Result{"test": {Details: []*Detail{{t: &date_1512085950_20171201085230, diff: &zero}, {t: &date_1512085960_20171201085240, diff: &ten}}, sum: &ten}},
			want: "Result:test : 2 times1 : 2017-12-01 08:52:30 +0900 JST : 0s2 : 2017-12-01 08:52:40 +0900 JST : 10saverage: 10 s",
		},
		{
			name: "Multiple events of multiple types",
			m:    map[string]*Result{"test": {Details: []*Detail{{t: &date_1512085950_20171201085230, diff: &zero}, {t: &date_1512085960_20171201085240, diff: &ten}}, sum: &ten}, "another": {Details: []*Detail{{t: &date_1512085970_20171201085250, diff: &zero}, {t: &date_1512085980_20171201085300, diff: &ten}}, sum: &ten}},
			want: "Result:test : 2 times1 : 2017-12-01 08:52:30 +0900 JST : 0s2 : 2017-12-01 08:52:40 +0900 JST : 10saverage: 10 sanother : 2 times1 : 2017-12-01 08:52:50 +0900 JST : 0s2 : 2017-12-01 08:53:00 +0900 JST : 10saverage: 10 s",
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
			gotPrint := strings.ReplaceAll(buf.String(), "\n", "")
			if gotPrint != tt.want {
				t.Errorf("print() = \n%v,\nwant \n%v", gotPrint, tt.want)
			}
		})
	}
}
