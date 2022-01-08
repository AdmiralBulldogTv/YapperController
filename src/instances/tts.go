package instance

import (
	"context"

	"github.com/admiralbulldogtv/yappercontroller/src/datastructures"
	"github.com/admiralbulldogtv/yappercontroller/src/textparser/parts"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TTS interface {
	SendRequest(ctx context.Context, text string, currentVoice parts.Voice, validVoices []parts.Voice, maxVoiceSwaps int) ([]byte, error)
	Generate(ctx context.Context, text string, id *primitive.ObjectID, channelID primitive.ObjectID, currentVoice parts.Voice, validVoices []parts.Voice, maxVoiceSwaps int, alert *datastructures.SseEventTtsAlert) error
	Skip(ctx context.Context, channelID primitive.ObjectID) error
	Reload(ctx context.Context, channelID primitive.ObjectID) error
}
