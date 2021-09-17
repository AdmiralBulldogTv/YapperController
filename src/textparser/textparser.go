package textparser

import (
	"context"
	"strings"

	"github.com/admiralbulldogtv/yappercontroller/src/global"
	"github.com/admiralbulldogtv/yappercontroller/src/textparser/currency"
	"github.com/admiralbulldogtv/yappercontroller/src/textparser/numbers"
	"github.com/admiralbulldogtv/yappercontroller/src/textparser/override"
	"github.com/admiralbulldogtv/yappercontroller/src/textparser/parts"
	"github.com/admiralbulldogtv/yappercontroller/src/textparser/sentance"
	"github.com/admiralbulldogtv/yappercontroller/src/textparser/strip"
	"github.com/admiralbulldogtv/yappercontroller/src/textparser/voice"
	"github.com/admiralbulldogtv/yappercontroller/src/textparser/words"
)

func Process(gCtx global.Context, ctx context.Context, text string) ([]parts.VoicePart, error) {
	mongo := gCtx.GetMongoInstance()

	voices, err := mongo.FetchVoices(ctx)
	if err != nil {
		return nil, err
	}

	validVoices := make([]parts.Voice, len(voices))
	for i, v := range voices {
		validVoices[i] = parts.Voice{
			Name:  v.Speaker,
			Type:  parts.VoicePartTypeReader,
			Entry: v,
		}
	}

	text = strings.ToLower(text)

	stat := override.NormalizeOverride([]parts.VoicePart{{Type: parts.PartTypeRaw, Value: text}})
	stat = voice.NormalizeVoices(stat, validVoices[0], validVoices)
	stat = currency.NormalizeCurrency(stat)
	stat = numbers.NormalizeNumbers(stat)
	stat = sentance.FixAbbreviations(stat)
	stat = words.NormalizeWords(stat)

	stat = sentance.SplitVoices(stat, 60)
	stat = strip.NormalizeCharacters(stat)

	return stat, nil
}
