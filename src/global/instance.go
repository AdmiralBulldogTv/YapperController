package global

import instance "github.com/admiralbulldogtv/yappercontroller/src/instances"

type Instance struct {
	Mongo instance.Mongo
	Redis instance.Redis
	TTS   instance.TTS
}
