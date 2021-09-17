package override

import (
	"regexp"
	"strings"

	"github.com/admiralbulldogtv/yappercontroller/src/textparser/parts"
)

var re = regexp.MustCompile(`[^a-z\s-|.,]`)

func NormalizeOverride(pts []parts.VoicePart) []parts.VoicePart {
	rPts := []parts.VoicePart{}
	for _, v := range pts {
		if v.Type != parts.PartTypeRaw {
			rPts = append(rPts, v)
			continue
		}
		text := v.Value
		firstIdx := -1
		for i := 0; i < len(text); i++ {
			switch text[i] {
			case '\\':
				i++
			case '`':
				if firstIdx == -1 {
					firstIdx = i
				} else {
					preText := text[0:firstIdx]
					if preText != "" {
						rPts = append(rPts, parts.VoicePart{Value: preText, Type: parts.PartTypeRaw})
					}
					rPts = append(rPts, parts.VoicePart{Value: re.ReplaceAllString(strings.ToLower(text[firstIdx+1:i]), ""), Type: parts.PartTypeOverride})
					text = text[i+1:]
					firstIdx = -1
					i = 0
				}
			}
		}
		if text != "" {
			rPts = append(rPts, parts.VoicePart{Value: text, Type: parts.PartTypeRaw})
		}
	}
	return rPts
}
