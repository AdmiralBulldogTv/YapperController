package tts

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"sync"
	"time"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
	"github.com/gobuffalo/packr/v2"
	"github.com/google/uuid"
	jsoniter "github.com/json-iterator/go"
	"github.com/orcaman/writerseeker"
	"github.com/sirupsen/logrus"
	"github.com/troydota/tts-textparser/src/datastructures"
	"github.com/troydota/tts-textparser/src/global"
	"github.com/troydota/tts-textparser/src/instances"
	"github.com/troydota/tts-textparser/src/textparser"
	"github.com/troydota/tts-textparser/src/textparser/parts"
	"github.com/troydota/tts-textparser/src/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type Request struct {
	Jid           string      `json:"jid"`
	Event         int         `json:"event"`
	ResponseEvent string      `json:"response_event"`
	Payload       interface{} `json:"payload"`
	Mode          int32       `json:"mode"`
}

type GenerateChangePayload struct {
	Speaker       string  `json:"speaker"`
	GPU           bool    `json:"gpu"`
	WarmUp        bool    `json:"warm_up"`
	GateThreshold float64 `json:"gate_threshold,omitempty"`
	Period        bool    `json:"period"`
	Start         int32   `json:"start"`
	Pace          float64 `json:"pace"`
	PitchShift    int32   `json:"pitch_shift"`
	PArpabet      float64 `json:"p_arpabet"`
	TacoPath      string  `json:"taco_path"`
	FastPath      string  `json:"fast_path"`
	OnnxPath      string  `json:"onnx_path"`
	CmuDictPath   string  `json:"cmudict_path"`

	Text string `json:"text"`
}

const (
	TTSModePrecise int32 = iota
	TTSModeFast
)

type Response struct {
	Event         int                    `json:"event"`
	Jid           string                 `json:"jid"`
	Wid           string                 `json:"wid"`
	Payload       GenerateChangeResponse `json:"payload"`
	ContentLength int                    `json:"content_length"`
	Time          float64                `json:"time"`
}

type GenerateChangeResponse struct {
	Data    string  `json:"data"`
	Length  float64 `json:"length"`
	Speaker string  `json:"speaker"`
}

type ttsInstance struct {
	gCtx        global.Context
	mp          map[string]chan string
	setKey      string
	outputEvent string
	mtx         sync.Mutex
	cb          map[string]chan Response
}

type respHelper struct {
	IdxMap map[int]parts.SpaceType
	Jid    string
	Resp   Response
	Voice  parts.Voice
	wav    *audio.IntBuffer
}

var (
	longPause   *audio.IntBuffer
	mediumPause *audio.IntBuffer
	shortPause  *audio.IntBuffer
)

func init() {
	var err error
	box := packr.New("tts-assets", "./assets/")
	lp, err := box.Find("long-pause.wav")
	if err != nil {
		panic(err)
	}
	mp, err := box.Find("medium-pause.wav")
	if err != nil {
		panic(err)
	}
	sp, err := box.Find("short-pause.wav")
	if err != nil {
		panic(err)
	}
	decoder := wav.NewDecoder(bytes.NewReader(lp))
	if !decoder.IsValidFile() {
		panic("bad file")
	}
	longPause, err = decoder.FullPCMBuffer()
	if err != nil {
		panic(err)
	}
	decoder = wav.NewDecoder(bytes.NewReader(mp))
	if !decoder.IsValidFile() {
		panic("bad file")
	}
	mediumPause, err = decoder.FullPCMBuffer()
	if err != nil {
		panic(err)
	}
	decoder = wav.NewDecoder(bytes.NewReader(sp))
	if !decoder.IsValidFile() {
		panic("bad file")
	}
	shortPause, err = decoder.FullPCMBuffer()
	if err != nil {
		panic(err)
	}
}

func NewInstance(ctx global.Context, setKey, outputEvent string) (instances.TtsInstance, error) {
	inst := &ttsInstance{
		gCtx:        ctx,
		mp:          make(map[string]chan string),
		setKey:      setKey,
		outputEvent: outputEvent,
		cb:          make(map[string]chan Response),
	}

	go inst.process()

	return inst, nil
}

func (inst *ttsInstance) process() {
	ch := make(chan string)
	inst.gCtx.GetRedisInstance().Subscribe(context.Background(), ch, inst.outputEvent)
	var (
		resp Response
		err  error
	)
	for msg := range ch {
		if err = json.UnmarshalFromString(msg, &resp); err != nil {
			panic(err)
		}
		inst.mtx.Lock()
		if v, ok := inst.cb[resp.Jid]; ok {
			v <- resp
			delete(inst.cb, resp.Jid)
		}
		inst.mtx.Unlock()
	}
}

func (inst *ttsInstance) SendRequest(ctx context.Context, text string) ([]byte, error) {
	_pts, err := textparser.Process(inst.gCtx, ctx, text)
	if err != nil {
		return nil, err
	}

	pts := parts.VoicePartList(_pts)

	cb := make(chan Response)
	results := map[string]*respHelper{}
	idxMap := map[int]*respHelper{}

	buf := &writerseeker.WriterSeeker{}
	buf.Reader()

	var encoder *wav.Encoder

	n := 0
	for voice, ptslist := range pts.Map() {
		for pt, meta := range ptslist.Unique() {
			jid, _ := uuid.NewRandom()
			rh := &respHelper{
				IdxMap: map[int]parts.SpaceType{},
				Jid:    jid.String(),
				Voice:  voice,
			}
			inst.mtx.Lock()
			if voice.Type == parts.VoicePartTypeReader {
				n++

				var (
					CmuDictPath string
					FastPath    string
					TacoPath    string
				)
				mode := TTSModePrecise
				if voice.Entry.FastPath != nil {
					mode = TTSModeFast
					CmuDictPath = *voice.Entry.CmuDictPath
					FastPath = *voice.Entry.FastPath
				} else {
					TacoPath = *voice.Entry.TacoPath
				}

				req, _ := json.MarshalToString(Request{
					Jid:           jid.String(),
					Event:         4,
					ResponseEvent: inst.outputEvent,
					Mode:          mode,
					Payload: GenerateChangePayload{
						Speaker:       voice.Entry.Speaker,
						GPU:           voice.Entry.GPU,
						WarmUp:        voice.Entry.WarmUp,
						GateThreshold: voice.Entry.GateThreshold,
						Period:        voice.Entry.Period,
						Start:         voice.Entry.Start,
						Pace:          voice.Entry.Pace,
						PitchShift:    voice.Entry.PitchShift,
						PArpabet:      voice.Entry.PArpabet,
						TacoPath:      TacoPath,
						FastPath:      FastPath,
						OnnxPath:      *voice.Entry.OnnxPath,
						CmuDictPath:   CmuDictPath,
						Text:          pt.Value,
					},
				})

				results[jid.String()] = rh
				inst.cb[jid.String()] = cb
				if err := inst.gCtx.GetRedisInstance().SAdd(ctx, inst.setKey, req); err != nil {
					return nil, err
				}
			}
			for _, v := range meta {
				idxMap[v.Idx] = rh
				rh.IdxMap[v.Idx] = v.Space
			}
			inst.mtx.Unlock()
		}
	}

	for i := 0; i < n; i++ {
		resp := <-cb
		results[resp.Jid].Resp = resp
		sDec, err := base64.StdEncoding.DecodeString(resp.Payload.Data)
		if err != nil {
			return nil, err
		}
		decoder := wav.NewDecoder(bytes.NewReader(sDec))
		if !decoder.IsValidFile() {
			return nil, err
		}
		if err != nil {
			return nil, err
		}
		aBuf, err := decoder.FullPCMBuffer()
		if err != nil {
			return nil, err
		}
		if encoder == nil {
			encoder = wav.NewEncoder(buf, aBuf.Format.SampleRate, int(decoder.BitDepth), aBuf.Format.NumChannels, int(decoder.WavAudioFormat))
		}
		results[resp.Jid].wav = aBuf
	}

	close(cb)

	// adding about 1 seconds on to each audio file.
	for i := 0; i < 5; i++ {
		if err = encoder.Write(longPause); err != nil {
			return nil, err
		}
	}

	for i := 0; i < len(idxMap); i++ {
		resp := idxMap[i]
		if resp.Voice.Type == parts.VoicePartTypeReader {
			if err := encoder.Write(resp.wav); err != nil {
				return nil, err
			}
		} else {
			// todo add sound bytes.
		}
		var arr *audio.IntBuffer
		switch resp.IdxMap[i] {
		case parts.SpaceTypeLongPause:
			arr = longPause
		case parts.SpaceTypeMediumPause:
			arr = mediumPause
		case parts.SpaceTypeShortPause:
			arr = shortPause
		default:
			logrus.Warnf("unknown pause %d", resp.IdxMap[i])
			continue
		}
		if arr != nil {
			if err = encoder.Write(arr); err != nil {
				return nil, err
			}
		}
	}

	// adding about 1 seconds on to each audio file.
	for i := 0; i < 5; i++ {
		if err = encoder.Write(longPause); err != nil {
			return nil, err
		}
	}

	if err := encoder.Close(); err != nil {
		return nil, err
	}

	if err := buf.Close(); err != nil {
		return nil, err
	}

	return ioutil.ReadAll(buf.Reader())
}

func (inst *ttsInstance) Generate(ctx context.Context, text string, id primitive.ObjectID, channelID primitive.ObjectID) ([]byte, error) {
	data, err := inst.SendRequest(ctx, text)
	if err != nil {
		return data, err
	}

	if err := inst.gCtx.GetRedisInstance().Set(ctx, fmt.Sprintf("generated:tts:%s", id.Hex()), utils.B2S(data), time.Minute); err != nil {
		return nil, err
	}

	if !channelID.IsZero() {
		event, err := json.MarshalToString(datastructures.SseEvent{
			Event: "tts",
			Payload: datastructures.SseEventTts{
				WavID: id,
			},
		})
		if err != nil {
			return nil, err
		}
		if err = inst.gCtx.GetRedisInstance().Publish(ctx, fmt.Sprintf("overlay:events:%s", channelID.Hex()), event); err != nil {
			return nil, err
		}
	}

	return data, err
}

func (inst *ttsInstance) Skip(ctx context.Context, channelID primitive.ObjectID) error {
	return inst.gCtx.GetRedisInstance().Publish(ctx, fmt.Sprintf("overlay:events:%s", channelID.Hex()), `{"event":"skip"}`)
}
