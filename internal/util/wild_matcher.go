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
	for i < len(str) && wild[i] != '*' {
		if wild[i] != str[i] && wild[i] != questionMark {
			return false
		}
		i++
	}

	j := i
	for i < len(str) {
		s := str[i]
		w := wild[j]
		if w == '*' {
			j++
			if j == len(wild) {
				return true
			}
			mp = wild[j:]
			cp = str[i+1:]
		} else if w == s || w == questionMark {
			i++
			j++
		} else {
			wild = mp
			j = 0

			str = cp
			cp = cp[1:]
			i = 0
		}
	}

	for wild[j] == '*' {
		j++
	}

	return j < len(wild)
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
	sps := strings.Split(wildcards, ",;")
	for _, name := range sps {
		if len(name) == 0 {
			continue
		}

		idx := strings.Index(name, pattern)
		if idx >= 0 {
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
