package sentance

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/jdkato/prose/v2"
	"github.com/troydota/tts-textparser/src/textparser/parts"
)

var reSpace = regexp.MustCompile(`\s|-`)

func SplitVoices(pts []parts.VoicePart, limit int) []parts.VoicePart {
	rPts := []parts.VoicePart{}
	for _, v := range pts {
		if v.Voice.Type == parts.VoicePartTypeByte {
			rPts = append(rPts, v)
			continue
		}
		var sentances []string
		doc, err := prose.NewDocument(v.Value)
		if err != nil {
			sentances = []string{reSpace.ReplaceAllString(v.Value, " ")}
		} else {
			sent := doc.Sentences()
			sentances = make([]string, len(sent))
			for i, s := range sent {
				sentances[i] = reSpace.ReplaceAllString(s.Text, " ")
			}
		}

		for _, s := range sentances {
			commas := strings.Split(s, ",")
			for i, cm := range commas {
				cm = strings.TrimSpace(cm)
				if len(cm) > limit {
					splits := strings.Split(cm, " ")
					currentBuild := ""
					for _, split := range splits {
						if len(currentBuild) != 0 && len(currentBuild)+len(split) > limit {
							rPts = append(rPts, parts.VoicePart{
								Value: currentBuild,
								Voice: v.Voice,
								VoicePartMeta: parts.VoicePartMeta{
									Space: parts.SpaceTypeShortPause,
								},
							})
							currentBuild = ""
						}
						currentBuild = strings.TrimSpace(currentBuild + " " + split)
					}
					if len(currentBuild) != 0 {
						rPts = append(rPts, parts.VoicePart{
							Value: currentBuild,
							Voice: v.Voice,
							VoicePartMeta: parts.VoicePartMeta{
								Space: parts.SpaceTypeShortPause,
							},
						})
					}
				} else {
					if i == len(commas)-1 {
						rPts = append(rPts, parts.VoicePart{
							Value: cm,
							Voice: v.Voice,
							VoicePartMeta: parts.VoicePartMeta{
								Space: parts.SpaceTypeLongPause,
							},
						})
					} else {
						rPts = append(rPts, parts.VoicePart{
							Value: cm,
							Voice: v.Voice,
							VoicePartMeta: parts.VoicePartMeta{
								Space: parts.SpaceTypeMediumPause,
							},
						})
					}
				}
			}
		}
	}

	return rPts
}

var abr = map[*regexp.Regexp]string{}

var rawAbr = map[string]string{
	"mrs":  "misess",
	"mr":   "mister",
	"mt":   "mount",
	"dr":   "doctor",
	"st":   "saint",
	"co":   "company",
	"jr":   "junior",
	"maj":  "major",
	"gen":  "general",
	"drs":  "doctors",
	"rev":  "reverend",
	"lt":   "lieutenant",
	"hon":  "honorable",
	"sgt":  "sergeant",
	"capt": "captain",
	"esq":  "esquire",
	"ltd":  "limited",
	"col":  "colonel",
	"ft":   "fort",
}

func init() {
	for k, v := range rawAbr {
		abr[regexp.MustCompile(fmt.Sprintf(`\b%s\.?\b`, k))] = v
	}
}

func FixAbbreviations(pts []parts.VoicePart) []parts.VoicePart {
	for i, v := range pts {
		txt := v.Value
		for re, repl := range abr {
			txt = re.ReplaceAllString(txt, repl)
		}
		v.Value = txt
		pts[i] = v
	}

	return pts
}
