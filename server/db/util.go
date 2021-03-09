package db

import (
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// SetPinyin sets the pinyin in parseable form
func SetPinyin(p string) string {
	segs := []string{}

	for _, oldSeg := range regexp.MustCompile("[\\p{L}\\p{Mn}]+\\p{N}?").FindAllString(p, -1) {
		newSeg := []rune{}
		tone := -1

		checkDia := func(c rune) {
			switch c {
			case '\u0300':
				tone = 4
			case '\u0301':
				tone = 2
			case '\u0305':
				tone = 1
			case '\u030c':
				tone = 3
			}
		}

		for _, c := range oldSeg {
			switch {
			case unicode.IsMark(c):
				checkDia(c)
			case unicode.IsNumber(c):
				if i, e := strconv.Atoi(string(c)); e == nil {
					tone = i
				}
			case unicode.IsLetter(c):
				newSeg = append(newSeg, c)
			}
		}

		t := transform.Chain(norm.NFD, transform.RemoveFunc(func(r rune) bool {
			state := unicode.Is(unicode.Mn, r)

			if state {
				checkDia(r)
			}

			return state
		}), norm.NFC)
		seg, _, _ := transform.String(t, string(newSeg))

		if tone == 0 {
			tone = 5
		}

		if tone != -1 {
			segs = append(segs, seg+"["+strconv.Itoa(tone)+"]")
		} else {
			segs = append(segs, seg)
		}
	}

	return strings.Join(segs, " ")
}

// MakePinyin makes pinyin in readable form
func MakePinyin(p string) string {
	segs := []string{}

	for _, seg := range strings.Split(p, " ") {
		if len(seg) > 3 && seg[len(seg)-1] == ']' && seg[len(seg)-3] == '[' {
			seg = seg[:len(seg)-3] + string(seg[len(seg)-2])
		}

		segs = append(segs, seg)
	}

	return strings.Join(segs, " ")
}
