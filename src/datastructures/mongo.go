package datastructures

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Overlay struct {
	ID        primitive.ObjectID `bson:"_id" json:"_id"`
	ChannelID primitive.ObjectID `bson:"channel_id" json:"channel_id"`
}

type Audio struct {
	ID        primitive.ObjectID `bson:"_id" json:"_id"`
	ChannelID primitive.ObjectID `bson:"channel_id" json:"channel_id"`
	Duration  time.Duration      `bson:"duration" json:"duration"`
	Segments  []AudioSegment     `bson:"segments" json:"segments"`
	Trigger   AudioTrigger       `bson:"trigger" json:"trigger"`
}

type AudioConfig struct {
	ID            primitive.ObjectID `bson:"_id" json:"_id"`
	Speaker       string             `bson:"speaker" json:"speaker"`
	GPU           bool               `bson:"gpu" json:"gpu"`
	WarmUp        bool               `bson:"warm_up" json:"warm_up"`
	GateThreshold float64            `bson:"gate_threshold" json:"gate_threshold"`
	Period        bool               `bson:"period" json:"period"`
	Start         int32              `bson:"start" json:"start"`
	Pace          float64            `bson:"pace" json:"pace"`
	PitchShift    int32              `bson:"pitch_shift" json:"pitch_shift"`
	PArpabet      float64            `bson:"p_arpabet" json:"p_arpabet"`
	Volume        float64            `bson:"volume" json:"volume"`
	TacoPath      *string            `bson:"taco_path" json:"taco_path"`
	FastPath      *string            `bson:"fast_path" json:"fast_path"`
	OnnxPath      *string            `bson:"onnx_path" json:"onnx_path"`
	CmuDictPath   *string            `bson:"cmudict_path" json:"cmudict_path"`
}

const (
	AudioConfigModePrecise int32 = iota
	AudioConfigModeFast
)

type AudioSegment struct {
	Voice     string        `bson:"voice" json:"voice"`
	Text      string        `bson:"text" json:"text"`
	StartTime time.Duration `bson:"start_time" json:"start_time"`
	Duration  time.Duration `bson:"duration" json:"duration"`
}

const (
	AudioTriggerSourceManual   = "MANUAL"
	AudioTriggerSourceSub      = "SUBSCRIPTION"
	AudioTriggerSourceDonation = "DONATION"
	AudioTriggerSourceBits     = "BITS"
)

type AudioTrigger struct {
	Source   string `bson:"source" json:"source"`
	Username string `bson:"username" json:"username"`
	Bits     int    `bson:"bits" json:"bits"`
	Amount   int    `bson:"donation" json:"donation"`
	Currency string `bson:"currency" json:"currency"`
}
