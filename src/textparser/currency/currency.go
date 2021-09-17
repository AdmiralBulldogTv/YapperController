package currency

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/admiralbulldogtv/yappercontroller/src/textparser/parts"
)

var re = regexp.MustCompile(`([-+])?\s*(?:([€£\$])\s*([-+])?\s*((?:[1-9]+0*)+(?:\s*,?\s*\d{3})*)?(?:\.(\d+))?|((?:[1-9]+0*)+(?:\s*,?\s*\d{3})*)?(?:\.(\d+))?\s*([€£\$]))`)
var numberRe = regexp.MustCompile(`[^\d]`)

var symbolMap = map[string]string{
	"€": "euro",
	"£": "pound",
	"$": "dollar",
}

var symbolMapCents = map[string]string{
	"€": "cents",
	"£": "pence",
	"$": "cents",
}

var symbolMapCent = map[string]string{
	"€": "cent",
	"£": "penny",
	"$": "cent",
}

func NormalizeCurrency(pts []parts.VoicePart) []parts.VoicePart {
	retParts := []parts.VoicePart{}
	for _, p := range pts {
		if p.Type != parts.PartTypeRaw || p.Value == "" {
			retParts = append(retParts, p)
			continue
		}

		text := strings.TrimSpace(p.Value)
		matches := re.FindAllStringSubmatch(text, -1)
		for _, match := range matches {
			neg := match[1]
			symbol := match[2]
			neg2 := match[3]
			if neg != "" && neg2 != "" {
				continue
			}
			if neg == "" {
				neg = neg2
			}
			if neg == "+" {
				neg = ""
			}
			number := match[4]
			decimal := match[5]
			if symbol == "" {
				number = match[6]
				decimal = match[7]
				symbol = match[8]
			}
			if number == "" && decimal == "" {
				continue
			}

			currency := symbolMap[symbol]
			number = numberRe.ReplaceAllString(number, "")
			if number != "1" {
				currency += "s"
			}

			var format string

			if decimal != "" {
				if len(decimal) <= 2 {
					if len(decimal) == 1 {
						decimal += "0"
					}
					cents := symbolMapCents[symbol]
					if decimal == "01" {
						cents = symbolMapCent[symbol]
					}
					if decimal[0] == '0' {
						decimal = decimal[1:]
					}
					decimal = fmt.Sprintf("%s %s", decimal, cents)

					if isZero(number) {
						format = decimal
					} else {
						format = fmt.Sprintf("%s %s and %s", number, currency, decimal)
					}
				} else if number == "" {
					format = fmt.Sprintf("0.%s %s", decimal, currency)
				} else {
					format = fmt.Sprintf("%s.%s %s", number, decimal, currency)
				}
			} else {
				format = fmt.Sprintf("%s %s", number, currency)
			}

			if neg == "-" {
				format = "minus " + format
			}

			mtch := strings.TrimSpace(match[0])

			idx := strings.Index(text, mtch)
			idx2 := idx + len(mtch)
			{
				if idx != 0 && text[idx-1] != ' ' {
					format = fmt.Sprint(" ", format)
				}
				if idx2 < len(text)-1 && text[idx2+1] != ' ' {
					format = fmt.Sprint(format, " ")
				}
			}

			preText := strings.TrimSpace(text[0:idx])
			text = strings.TrimSpace(text[idx2:])
			if preText != "" {
				for symbol, replace := range symbolMap {
					preText = replaceAddSpace(preText, symbol, replace)
				}
				retParts = append(retParts, parts.VoicePart{Type: parts.PartTypeRaw, Value: preText, Voice: p.Voice, VoicePartMeta: p.VoicePartMeta})
			}
			retParts = append(retParts, parts.VoicePart{Type: parts.PartTypeCurrency, Value: format, Voice: p.Voice, VoicePartMeta: p.VoicePartMeta})
		}
		if text != "" {
			for symbol, replace := range symbolMap {
				text = replaceAddSpace(text, symbol, replace)
			}
			retParts = append(retParts, parts.VoicePart{Type: parts.PartTypeRaw, Value: text, Voice: p.Voice, VoicePartMeta: p.VoicePartMeta})
		}
	}

	return retParts
}

func replaceAddSpace(text, old, new string) string {
	edited := text
	idx := 0
	for {
		idx = strings.Index(text[idx:], old)
		if idx == -1 {
			break
		}
		repl := new
		if idx != 0 && text[idx-1] != ' ' {
			repl = " " + new
		}
		if idx+1 < len(text) && text[idx+1] != ' ' {
			repl = new + " "
		}
		edited = strings.Replace(edited, old, repl, 1)
		idx++
	}
	return edited
}

func isZero(number string) bool {
	for _, v := range number {
		if v != '0' {
			return false
		}
	}
	return true
}
