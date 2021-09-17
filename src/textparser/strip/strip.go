package strip

import (
	"regexp"

	"github.com/admiralbulldogtv/yappercontroller/src/textparser/parts"
)

var re = regexp.MustCompile(`[^a-z |]`)
var reSpace = regexp.MustCompile(`\s|-`)
var spaceFix = regexp.MustCompile(`\s+`)

func NormalizeCharacters(pts []parts.VoicePart) []parts.VoicePart {
	for i := range pts {
		pts[i].Value = spaceFix.ReplaceAllString(re.ReplaceAllString(reSpace.ReplaceAllString(pts[i].Value, " "), ""), " ")
	}

	return pts
}
