package numbers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/admiralbulldogtv/yappercontroller/src/textparser/parts"
)

var re = regexp.MustCompile(`([-+])?(\d+(?:\s*,?\s*\d{3})*)(?:\.(\d+))?`)
var numberRe = regexp.MustCompile(`[^\d]`)

var tens = []string{"", "ten", "twenty", "thirty", "fourty", "fifty", "sixty", "seventy", "eighty", "ninety"}
var units = []string{"zero", "one", "two", "three", "four", "five", "six", "seven", "eight", "nine", "ten", "eleven", "twelve", "thirteen", "fourteen", "fifteen", "sixteen", "seventeen", "eighteen", "nineteen"}

var dividers = []string{
	"",
	"thousand",
	"million",
	"billion",
	"trillion",
	"quadrillion",
	"quintillion",
	"sextillion",
	"septillion",
	"octillion",
	"nonillion",
	"decillion",
	"undecillion",
	"duodecillion",
	"tredecillion",
	"quattuordecillion",
	"quindecillion",
	"sexdecillion",
	"septendecillion",
	"octodecillion",
	"novemdecillion",
	"vigintillion",
	"unvigintillion",
	"dovigintillion",
	"trevigintillion",
	"quattuorvigintillion",
	"quinvigintillion",
	"sexvigintillion",
	"septenvigintillion",
	"octovigintillion",
	"novemvigintillion",
	"trigintillion",
	"untrigintillion",
	"dotrigintillion",
	"tretrigintillion",
	"quattuortrigintillion",
	"quintrigintillion",
	"sextrigintillion",
	"septentrigintillion",
	"octotrigintillion",
	"novemtrigintillion",
}

func NormalizeNumbers(pts []parts.VoicePart) []parts.VoicePart {
	retParts := []parts.VoicePart{}
	for _, p := range pts {
		switch p.Type {
		case parts.PartTypeCurrency, parts.PartTypeRaw:
		default:
			retParts = append(retParts, p)
			continue
		}
		text := strings.TrimSpace(p.Value)
		matches := re.FindAllStringSubmatch(text, -1)
		p.Value = ""
		for _, match := range matches {
			idx := strings.Index(text, match[0])
			idx2 := idx + len(match[0])

			sign := match[1]
			number := numberRe.ReplaceAllString(match[2], "")
			preText := strings.TrimSpace(text[0:idx])
			text = strings.TrimSpace(text[idx2:])

			if number[0] == '0' {
				number = strings.Join(numberDigits(number), " ")
			} else {
				number = normalizeNumber(number)
			}

			decimal := match[3]
			if decimal != "" {
				numbers := numberDigits(decimal)
				if len(numbers) != 0 {
					decimal = fmt.Sprint("point ", strings.Join(numbers, " "))
				} else {
					decimal = ""
				}
			}

			if sign == "-" {
				sign = "minus"
			} else {
				sign = ""
			}

			p.Value = strings.TrimSpace(fmt.Sprint(p.Value, " ", preText, " ", strings.TrimSpace(fmt.Sprint(sign, " ", number)), " ", decimal))
		}
		if text != "" {
			p.Value += " " + text
		}
		p.Value = strings.TrimSpace(p.Value)
		retParts = append(retParts, p)
	}

	return retParts
}

func normalizeNumber(number string) string {
	numbers := splitNumbers(number)
	var parts []string
	l := len(numbers)
	if l > len(dividers) {
		// this number is too big to read so we should read the digits.
		parts = numberDigits(number)
	} else {
		parts = make([]string, l)
		for i, v := range numbers {
			parts[l-i-1] = fmt.Sprint(convertNumber(v), " ", dividers[i])
		}
	}
	return strings.TrimSpace(strings.Join(parts, ", "))
}

func splitNumbers(number string) []int {
	l := len(number)
	splits := []int{}
	currentNum := ""
	for i := 0; i < l; i++ {
		if i != 0 && i%3 == 0 {
			n, err := strconv.Atoi(currentNum)
			if err != nil {
				panic(err)
			}
			splits = append(splits, n)
			currentNum = ""
		}
		currentNum = string(number[l-i-1]) + currentNum
	}
	if currentNum != "" {
		n, err := strconv.Atoi(currentNum)
		if err != nil {
			panic(err)
		}
		splits = append(splits, n)
	}

	return splits
}

func convertNumber(number int) string {
	parts := []string{}
	if number >= 100 {
		n := number / 100
		number -= n * 100
		parts = append(parts, units[n], "hundred")
	}
	if number >= 20 {
		n := number / 10
		number -= n * 10
		if len(parts) != 0 {
			parts = append(parts, "and")
		}
		parts = append(parts, tens[n])
	}
	if number != 0 {
		parts = append(parts, units[number])
	}

	return strings.Join(parts, " ")
}

func numberDigits(number string) []string {
	parts := make([]string, len(number))
	for i, r := range number {
		n, err := strconv.Atoi(string(r))
		if err != nil {
			panic(err)
		}
		parts[i] = units[n]
	}
	return parts
}
