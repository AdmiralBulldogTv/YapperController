package datastructures

import "go.mongodb.org/mongo-driver/bson/primitive"

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
}

type SseEventTtsTranscription struct {
	Voice    string  `json:"voice"`
	Duration float64 `json:"duration"`
	Text     string  `json:"text"`
}
