package datastructures

import "go.mongodb.org/mongo-driver/bson/primitive"

type SseEvent struct {
	Event   string      `json:"event"`
	Payload interface{} `json:"payload"`
}

type SseEventTts struct {
	WavID          primitive.ObjectID         `json:"wav_id"`
	Transcriptions []SseEventTtsTranscription `json:"transcriptions"`
}

type SseEventTtsTranscription struct {
	Voice    string  `json:"voice"`
	Duration float64 `json:"duration"`
	Text     string  `json:"text"`
}
