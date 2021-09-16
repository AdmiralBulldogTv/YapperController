package datastructures

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Overlay struct {
	ID        primitive.ObjectID `bson:"_id"`
	ChannelID primitive.ObjectID `bson:"channel_id"`
}

type Audio struct {
	ID        primitive.ObjectID `bson:"_id"`
	ChannelID primitive.ObjectID `bson:"channel_id"`
	Duration  time.Duration      `bson:"duration"`
	Segments  []AudioSegment     `bson:"segments"`
	Trigger   AudioTrigger       `bson:"trigger"`
}

type AudioConfig struct {
	ID            primitive.ObjectID `bson:"_id"`
	Speaker       string             `bson:"speaker"`
	GPU           bool               `bson:"gpu"`
	WarmUp        bool               `bson:"warm_up"`
	GateThreshold float64            `bson:"gate_threshold"`
	Period        bool               `bson:"period"`
	Start         int32              `bson:"start"`
	Pace          float64            `bson:"pace"`
	PitchShift    int32              `bson:"pitch_shift"`
	PArpabet      float64            `bson:"p_arpabet"`
	TacoPath      *string            `bson:"taco_path"`
	FastPath      *string            `bson:"fast_path"`
	OnnxPath      *string            `bson:"onnx_path"`
	CmuDictPath   *string            `bson:"cmudict_path"`
}

const (
	AudioConfigModePrecise int32 = iota
	AudioConfigModeFast
)

type AudioSegment struct {
	Voice     string        `bson:"voice"`
	Text      string        `bson:"text"`
	StartTime time.Duration `bson:"start_time"`
	Duration  time.Duration `bson:"duration"`
}

const (
	AudioTriggerSourceManual   = "MANUAL"
	AudioTriggerSourceSub      = "SUBSCRIPTION"
	AudioTriggerSourceDonation = "DONATION"
	AudioTriggerSourceBits     = "BITS"
)

type AudioTrigger struct {
	Source   string `bson:"source"`
	Username string `bson:"username"`
	Bits     int    `bson:"bits"`
	Amount   int    `bson:"donation"`
	Currency string `bson:"currency"`
}
