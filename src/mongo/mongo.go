package mongo

import (
	"context"

	"github.com/admiralbulldogtv/yappercontroller/src/datastructures"
	"github.com/admiralbulldogtv/yappercontroller/src/instances"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mongoInstance struct {
	c  *mongo.Client
	db *mongo.Database
}

func (i *mongoInstance) Ping(ctx context.Context) error {
	return i.c.Ping(ctx, nil)
}

func (i *mongoInstance) FetchOverlay(ctx context.Context, token primitive.ObjectID) (datastructures.Overlay, error) {
	o := datastructures.Overlay{}
	res := i.db.Collection("overlays").FindOne(ctx, bson.M{"_id": token})
	err := res.Err()
	if err == nil {
		err = res.Decode(&o)
	}
	return o, err
}

func (i *mongoInstance) FetchVoices(ctx context.Context) ([]datastructures.AudioConfig, error) {
	vcs := []datastructures.AudioConfig{}
	cur, err := i.db.Collection("audio_configs").Find(ctx, bson.M{})
	if err == nil {
		err = cur.All(ctx, &vcs)
	}
	return vcs, err
}

func NewInstance(ctx context.Context, uri, db string) (instances.MongoInstance, error) {
	c, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	i := &mongoInstance{
		c:  c,
		db: c.Database(db),
	}

	if err = c.Connect(ctx); err != nil {
		return nil, err
	}

	if err = i.Ping(ctx); err != nil {
		return nil, err
	}

	return i, nil
}
