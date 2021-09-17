package words

import (
	"regexp"
	"strings"

	"github.com/admiralbulldogtv/yappercontroller/src/textparser/parts"
)

var re = regexp.MustCompile(`\s*([.,]+)\s*`)

func NormalizeWords(pts []parts.VoicePart) []parts.VoicePart {
	for i, v := range pts {
		switch v.Type {
		case parts.PartTypeCurrency, parts.PartTypeRaw:
		default:
			continue
		}
		words := strings.Split(re.ReplaceAllString(v.Value, " $1 "), " ")
		for i, word := range words {
			if v, ok := wordLetterMap[word]; ok {
				spl := strings.Split(v, " ")
				for i, sp := range spl {
					n, ok := wordMap[sp]
					if !ok {
						n = sp
					}
					spl[i] = n
				}
				words[i] = strings.Join(spl, " ")
			} else {
				if v, ok = wordMap[word]; !ok {
					if len(word) > 20 {
						spl := make([]string, len(word))
						for i, k := range word {
							spl[i] = wordMap[string(k)]
						}
						v = strings.Join(spl, " ")
					} else {
						v = word
					}
				}
				words[i] = v
			}
		}
		v.Value = strings.Join(words, " ")
		pts[i] = v
	}
	return pts
}

var wordMap = map[string]string{
	"totsugeki":      "tot|sue|geck|ki",
	"bulldog":        "bull|dog",
	"omegalul":       "oh|meg|ga|lul",
	"a":              "ay",
	"b":              "bee",
	"c":              "see",
	"d":              "dee",
	"e":              "ee",
	"f":              "eff",
	"g":              "gee",
	"h":              "aych",
	"j":              "jay",
	"k":              "kay",
	"l":              "el",
	"m":              "em",
	"n":              "en",
	"o":              "oh",
	"p":              "pee",
	"q":              "kyoo",
	"r":              "ahr",
	"s":              "ess",
	"t":              "tee",
	"u":              "you",
	"v":              "vee",
	"w":              "double|you",
	"x":              "eks",
	"y":              "why",
	"z":              "zee",
	"xd":             "eks|de",
	"lacari":         "la|car|ee",
	"fucking":        "fucking",
	"+":              "plus",
	"..":             "dot dot",
	"...":            "dot dot dot",
	"homie":          "home|e",
	"homies":         "home|ees",
	"ez":             "ease",
	"batchest":       "bat|chest",
	"hecking":        "hec|king",
	"pepelaugh":      "pep|ay|laugh",
	"kekw":           "kek double|you",
	"kekl":           "kek el",
	"kekinsane":      "kek insane",
	"and":            "aand",
	"wtf":            "what the fuck",
	"admiralbulldog": "admiral bull|dog",

	// currency terms
	"thousand":              "thousand",
	"million":               "million",
	"billion":               "billion",
	"trillion":              "trillion",
	"quadrillion":           "quadrillion",
	"quintillion":           "quintillion",
	"sextillion":            "sextillion",
	"septillion":            "septillion",
	"octillion":             "octillion",
	"nonillion":             "nonillion",
	"decillion":             "decillion",
	"undecillion":           "undecillion",
	"duodecillion":          "duodecillion",
	"tredecillion":          "tredecillion",
	"quattuordecillion":     "quattuordecillion",
	"quindecillion":         "quindecillion",
	"sexdecillion":          "sexdecillion",
	"septendecillion":       "septendecillion",
	"octodecillion":         "octodecillion",
	"novemdecillion":        "novemdecillion",
	"vigintillion":          "vigintillion",
	"unvigintillion":        "unvigintillion",
	"dovigintillion":        "dovigintillion",
	"trevigintillion":       "trevigintillion",
	"quattuorvigintillion":  "quattuorvigintillion",
	"quinvigintillion":      "quinvigintillion",
	"sexvigintillion":       "sexvigintillion",
	"septenvigintillion":    "septenvigintillion",
	"octovigintillion":      "octovigintillion",
	"novemvigintillion":     "novemvigintillion",
	"trigintillion":         "trigintillion",
	"untrigintillion":       "untrigintillion",
	"dotrigintillion":       "dotrigintillion",
	"tretrigintillion":      "tretrigintillion",
	"quattuortrigintillion": "quattuortrigintillion",
	"quintrigintillion":     "quintrigintillion",
	"sextrigintillion":      "sextrigintillion",
	"septentrigintillion":   "septentrigintillion",
	"octotrigintillion":     "octotrigintillion",
	"novemtrigintillion":    "novemtrigintillion",
}

var wordLetterMap = map[string]string{
	"abc": "a b c",
	"xyz": "x y z",
	"bkb": "b k b",
	"tb":  "t b",
	"mkb": "m k b",
	"fml": "f m l",
	"idk": "i d k",
	"og":  "o g",
	"eg":  "e g",
	"tts": "t t s",
	"bf":  "b f",
}
