package textparser

import (
	"context"
	"strings"

	"github.com/troydota/tts-textparser/src/global"
	"github.com/troydota/tts-textparser/src/textparser/currency"
	"github.com/troydota/tts-textparser/src/textparser/numbers"
	"github.com/troydota/tts-textparser/src/textparser/override"
	"github.com/troydota/tts-textparser/src/textparser/parts"
	"github.com/troydota/tts-textparser/src/textparser/sentance"
	"github.com/troydota/tts-textparser/src/textparser/strip"
	"github.com/troydota/tts-textparser/src/textparser/voice"
	"github.com/troydota/tts-textparser/src/textparser/words"
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
