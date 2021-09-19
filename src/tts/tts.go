package tts

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path"
	"sync"
	"time"

	"github.com/admiralbulldogtv/yappercontroller/src/datastructures"
	"github.com/admiralbulldogtv/yappercontroller/src/global"
	"github.com/admiralbulldogtv/yappercontroller/src/instances"
	"github.com/admiralbulldogtv/yappercontroller/src/textparser"
	"github.com/admiralbulldogtv/yappercontroller/src/textparser/parts"
	"github.com/admiralbulldogtv/yappercontroller/src/utils"
	"github.com/gobuffalo/packr/v2"
	"github.com/google/uuid"
	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
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
}

var (
	soundMap = map[string][]byte{}
)

func init() {
	box := packr.New("tts-assets", "./assets/")
	files := []string{"long-pause.wav", "medium-pause.wav", "short-pause.wav"}
	for _, v := range files {
		sp, err := box.Find(v)
		if err != nil {
			panic(err)
		}
		soundMap[v] = sp
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

func (inst *ttsInstance) SendRequest(ctx context.Context, text string, currentVoice parts.Voice, validVoices []parts.Voice, maxVoiceSwaps int) ([]byte, error) {
	_pts, err := textparser.Process(text, currentVoice, validVoices, maxVoiceSwaps)
	if err != nil {
		return nil, err
	}

	pts := parts.VoicePartList(_pts)

	cb := make(chan Response)
	results := map[string]*respHelper{}
	idxMap := map[int]*respHelper{}

	tmpPath := path.Join("tmp", uuid.NewString())
	if err = os.MkdirAll(tmpPath, 0700); err != nil {
		return nil, err
	}

	defer os.RemoveAll(tmpPath)

	files := []string{}

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
		if err = os.WriteFile(path.Join(tmpPath, fmt.Sprintf("%s.wav", resp.Jid)), sDec, 0666); err != nil {
			return nil, err
		}
	}

	close(cb)

	used := map[string]bool{}

	for i := 0; i < len(idxMap); i++ {
		resp := idxMap[i]
		if resp.Voice.Type == parts.VoicePartTypeReader {
			files = append(files, path.Join(tmpPath, fmt.Sprintf("%s.wav", resp.Jid)))
		}
		// else {
		// 	// todo add sound bytes.
		// }
		var pause string
		switch resp.IdxMap[i] {
		case parts.SpaceTypeLongPause:
			pause = "medium-pause.wav"
		case parts.SpaceTypeMediumPause:
			pause = "short-pause.wav"
		case parts.SpaceTypeShortPause:
		default:
			logrus.Warnf("unknown pause %d", resp.IdxMap[i])
			continue
		}
		if pause != "" {
			pth := path.Join(tmpPath, pause)
			if !used[pause] {
				if err = os.WriteFile(pth, soundMap[pause], 0666); err != nil {
					return nil, err
				}
				used[pause] = true
			}
			files = append(files, pth)
		}
	}

	outPth := path.Join(tmpPath, "output.wav")

	files = append(files, "-c", "2", "-r", "48000", outPth)

	if err = exec.CommandContext(ctx, "sox", files...).Run(); err != nil {
		return nil, err
	}

	return os.ReadFile(outPth)
}

func (inst *ttsInstance) Generate(ctx context.Context, text string, id primitive.ObjectID, channelID primitive.ObjectID, currentVoice parts.Voice, validVoices []parts.Voice, maxVoiceSwaps int) ([]byte, error) {
	data, err := inst.SendRequest(ctx, text, currentVoice, validVoices, maxVoiceSwaps)
	if err != nil {
		return data, err
	}

	if err := inst.gCtx.GetRedisInstance().Set(ctx, fmt.Sprintf("generated:tts:%s", id.Hex()), utils.B2S(data), time.Hour); err != nil {
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
