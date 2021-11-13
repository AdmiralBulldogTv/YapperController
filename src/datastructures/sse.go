package datastructures

import (
	"github.com/admiralbulldogtv/yappercontroller/src/alerts"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SseEvent struct {
	Event   string      `json:"event"`
	Payload interface{} `json:"payload"`
}

type SseEventTts struct {
	WavID *primitive.ObjectID `json:"wav_id"`
	Alert *SseEventTtsAlert   `json:"alert"`
}

type SseEventTtsAlert struct {
	Type    string `json:"type"`
	Image   string `json:"image"`
	Audio   string `json:"audio"`
	Text    string `json:"text"`
	SubText string `json:"sub_text"`
	Volume  int    `json:"volume,omitempty"`
}

type SseEventTtsTranscription struct {
	Voice    string  `json:"voice"`
	Duration float64 `json:"duration"`
	Text     string  `json:"text"`
}

type AlertHelper struct {
	Type string
	Name string
}

func (a AlertHelper) Parse() (string, string, int) {
	switch a.Type {
	case "cheer":
		return alerts.CheerAlerts[a.Name+".gif"].ToName(), alerts.CheerAlerts[a.Name+".wav"].ToName(), alerts.CheerAlerts[a.Name+".wav"].Volume
	case "donation":
		return alerts.DonationAlerts[a.Name+".gif"].ToName(), alerts.DonationAlerts[a.Name+".wav"].ToName(), alerts.DonationAlerts[a.Name+".wav"].Volume
	case "subscriber":
		return alerts.SubscriberAlerts[a.Name+".gif"].ToName(), alerts.SubscriberAlerts[a.Name+".wav"].ToName(), alerts.SubscriberAlerts[a.Name+".wav"].Volume
	default:
		return "", "", 0
	}
}
