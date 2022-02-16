package util

import (
	"strings"
)

type WildcardMatcher struct {
	questionMark byte
	onBothEmpty  bool
	wcs          []string
	exact        []string
}

func contains(a []string, x string) bool {
	for _, v := range a {
		if v == x {
			return true
		}
	}
	return false
}

/*
inline bool wildcmp(const char* wild, const char* str, char question_mark) {
    const char* cp = NULL;
    const char* mp = NULL;

    while (*str && *wild != '*') {
        if (*wild != *str && *wild != question_mark) {
            return false;
        }
        ++wild;
        ++str;
    }

    while (*str) {
        if (*wild == '*') {
            if (!*++wild) {
                return true;
            }
            mp = wild;
            cp = str+1;
        } else if (*wild == *str || *wild == question_mark) {
            ++wild;
            ++str;
        } else {
            wild = mp;
            str = cp++;
        }
    }

    while (*wild == '*') {
        ++wild;
    }
    return !*wild;
}

*/
func wildcmp(wild string, str string, questionMark byte) bool {
	var (
		cp string
		mp string
	)

	i := 0
	for i < len(str) && i < len(wild) && wild[i] != '*' {
		if wild[i] != str[i] && wild[i] != questionMark {
			return false
		}
		i++
	}

	wild = wild[i:]
	str = str[i:]
	for len(str) > 0 {
		s := str[0]
		w := uint8(0)
		if len(wild) > 0 {
			w = wild[0]
		}
		if w == '*' {
			wild = wild[1:]
			if len(wild) == 0 {
				return true
			}
			mp = wild
			cp = str[1:]
		} else if w == s || w == questionMark {
			str = str[1:]
			wild = wild[1:]
		} else {
			wild = mp
			str = cp
			if len(cp) > 0 {
				cp = cp[1:]
			}
		}
	}

	for len(wild) > 0 && wild[0] == '*' {
		wild = wild[1:]
	}

	return len(wild) == 0
}
func (w *WildcardMatcher) Wildcards() []string {
	return w.wcs
}
func (w *WildcardMatcher) ExactNames() []string {
	return w.exact
}
func (w *WildcardMatcher) Match(name string) bool {
	if len(w.exact) != 0 {
		if contains(w.exact, name) {
			return true
		}
	} else if len(w.wcs) == 0 {
		return w.onBothEmpty
	}

	for _, v := range w.wcs {
		if wildcmp(v, name, w.questionMark) {
			return true
		}
	}

	return false
}
func splitAny(s string, seps string) []string {
	splitter := func(r rune) bool {
		return strings.ContainsRune(seps, r)
	}
	return strings.FieldsFunc(s, splitter)
}
func NewWildcardMatcher(wildcards string, questionMark byte, onBothEmpty bool) *WildcardMatcher {
	w := &WildcardMatcher{
		questionMark: questionMark,
		onBothEmpty:  onBothEmpty,
		wcs:          nil,
		exact:        nil,
	}

	if len(wildcards) == 0 {
		return w
	}

	pattern := string([]byte{'*', questionMark})
	//sps := strings.Split(wildcards, ",;")
	sps := splitAny(wildcards, ",;")
	for _, name := range sps {
		if len(name) == 0 {
			continue
		}

		if strings.ContainsAny(name, pattern) {
			w.wcs = append(w.wcs, name)
		} else {
			found := false
			for _, v := range w.exact {
				if v == name {
					found = true
					break
				}
			}
			if !found {
				w.exact = append(w.exact, name)
			}
		}
	}

	return w
}
