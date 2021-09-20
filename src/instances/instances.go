package instances

import (
	"context"
	"time"

	"github.com/admiralbulldogtv/yappercontroller/src/datastructures"
	"github.com/admiralbulldogtv/yappercontroller/src/textparser/parts"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TtsInstance interface {
	SendRequest(ctx context.Context, text string, currentVoice parts.Voice, validVoices []parts.Voice, maxVoiceSwaps int) ([]byte, error)
	Generate(ctx context.Context, text string, id *primitive.ObjectID, channelID primitive.ObjectID, currentVoice parts.Voice, validVoices []parts.Voice, maxVoiceSwaps int, alert *datastructures.SseEventTtsAlert) error
	Skip(ctx context.Context, channelID primitive.ObjectID) error
	Reload(ctx context.Context, channelID primitive.ObjectID) error
}

type RedisInstance interface {
	Ping(ctx context.Context) error
	Subscribe(ctx context.Context, ch chan string, subscribeTo ...string)
	Publish(ctx context.Context, channel string, data string) error
	SAdd(ctx context.Context, set string, values ...interface{}) error
	Set(ctx context.Context, key string, value string, expiry time.Duration) error
	Get(ctx context.Context, key string) (string, error)
}

type MongoInstance interface {
	Ping(ctx context.Context) error
	FetchOverlay(ctx context.Context, token primitive.ObjectID) (datastructures.Overlay, error)
	FetchVoices(ctx context.Context) ([]datastructures.AudioConfig, error)
}
