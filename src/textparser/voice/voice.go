package voice

import (
	"strings"

	"github.com/admiralbulldogtv/yappercontroller/src/textparser/parts"
)

func NormalizeVoices(pts []parts.VoicePart, currentVoice parts.Voice, validVoices []parts.Voice) []parts.VoicePart {
	returnPts := []parts.VoicePart{}

	// spew.Dump(pts)

	for _, part := range pts {
		currentBuild := []string{}
		splits := strings.Split(part.Value, " ")
		for _, v := range splits {
			found := false
			for _, voice := range validVoices {
				switch voice.Type {
				case parts.VoicePartTypeByte:
					if strings.HasPrefix(v, "("+voice.Name+")") {
						if len(currentBuild) != 0 {
							txt := strings.TrimSpace(strings.Join(currentBuild, " "))
							if len(txt) != 0 {
								returnPts = append(returnPts, parts.VoicePart{Value: txt, Voice: currentVoice, Type: part.Type})
							}
							currentBuild = []string{}
						}
						returnPts = append(returnPts, parts.VoicePart{Value: "", Voice: voice, Type: part.Type})
						found = true
					}
				case parts.VoicePartTypeReader:
					if strings.HasPrefix(v, voice.Name+":") {
						if voice != currentVoice {
							if len(currentBuild) != 0 {
								txt := strings.TrimSpace(strings.Join(currentBuild, " "))
								if len(txt) != 0 {
									returnPts = append(returnPts, parts.VoicePart{Value: txt, Voice: currentVoice, Type: part.Type})
								}
								currentBuild = []string{}
							}
							currentVoice = voice
						}
						currentBuild = append(currentBuild, strings.TrimPrefix(v, voice.Name+":"))
						found = true
					}
				}
				if found {
					break
				}
			}
			if !found {
				currentBuild = append(currentBuild, v)
			}
		}
		if len(currentBuild) != 0 {
			returnPts = append(returnPts, parts.VoicePart{Value: strings.TrimSpace(strings.Join(currentBuild, " ")), Voice: currentVoice, Type: part.Type})
		}
	}

	return returnPts
}
