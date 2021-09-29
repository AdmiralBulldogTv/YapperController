package textparser

import (
	"strings"

	"github.com/admiralbulldogtv/yappercontroller/src/datastructures"
	"github.com/admiralbulldogtv/yappercontroller/src/textparser/currency"
	"github.com/admiralbulldogtv/yappercontroller/src/textparser/numbers"
	"github.com/admiralbulldogtv/yappercontroller/src/textparser/override"
	"github.com/admiralbulldogtv/yappercontroller/src/textparser/parts"
	"github.com/admiralbulldogtv/yappercontroller/src/textparser/sentance"
	"github.com/admiralbulldogtv/yappercontroller/src/textparser/strip"
	"github.com/admiralbulldogtv/yappercontroller/src/textparser/voice"
	"github.com/admiralbulldogtv/yappercontroller/src/textparser/words"
	"github.com/gobuffalo/packr/v2"
	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

var Voices []parts.Voice
var VoicesMap map[string]parts.Voice = map[string]parts.Voice{}

func init() {
	box := packr.New("textparser-static", "./static")
	data, err := box.Find("configs.json")
	if err != nil {
		panic(err)
	}

	cfgs := []datastructures.AudioConfig{}

	if err := json.Unmarshal(data, &cfgs); err != nil {
		panic(err)
	}

	Voices = make([]parts.Voice, len(cfgs))
	for i, v := range cfgs {
		Voices[i] = parts.Voice{
			Name:  v.Speaker,
			Type:  parts.VoicePartTypeReader,
			Entry: v,
		}
	}

	for _, v := range Voices {
		VoicesMap[v.Name] = v
	}
}

func Process(text string, currentVoice parts.Voice, validVoices []parts.Voice, maxVoiceSwaps int) ([]parts.VoicePart, error) {
	text = strings.ToLower(text)

	stat := override.NormalizeOverride([]parts.VoicePart{{Type: parts.PartTypeRaw, Value: text}})
	stat = voice.NormalizeVoices(stat, currentVoice, validVoices, maxVoiceSwaps)
	stat = currency.NormalizeCurrency(stat)
	stat = numbers.NormalizeNumbers(stat)
	stat = sentance.FixAbbreviations(stat)
	stat = words.NormalizeWords(stat)

	stat = sentance.SplitVoices(stat, 250)
	stat = strip.NormalizeCharacters(stat)

	return stat, nil
}
