package instance

import (
	"context"

	"github.com/admiralbulldogtv/yappercontroller/src/datastructures"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Mongo interface {
	Ping(ctx context.Context) error
	FetchOverlay(ctx context.Context, token primitive.ObjectID) (datastructures.Overlay, error)
	FetchVoices(ctx context.Context) ([]datastructures.AudioConfig, error)
}
