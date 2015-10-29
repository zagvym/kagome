package splitter

import (
	"unicode"
	"unicode/utf8"
)

// SentenceSplitter is a tiny sentence splitter for japanese texts.
type SentenceSplitter struct {
	Delim               []rune
	Follower            []rune
	SkipWhiteSpace      bool
	DoubleLineFeedSplit bool
	MaxRuneLen          int
}

var (
	// default sentence splitter
	splitter = &SentenceSplitter{
		Delim:               []rune{'。', '．'},
		Follower:            []rune{'」', '』'},
		SkipWhiteSpace:      true,
		DoubleLineFeedSplit: true,
		MaxRuneLen:          256,
	}
)

// ScanSentences is a split function for a Scanner that returns each sentece of text.
func ScanSentences(data []byte, atEOF bool) (advance int, token []byte, err error) {
	return splitter.ScanSentences(data, atEOF)
}

func (s SentenceSplitter) isDelim(r rune) bool {
	for _, d := range s.Delim {
		if r == d {
			return true
		}
	}
	return false
}

func (s SentenceSplitter) isFollower(r rune) bool {
	for _, d := range s.Follower {
		if r == d {
			return true
		}
	}
	return false
}

// ScanSentences is a split function for a Scanner that returns each sentece of text.
func (s SentenceSplitter) ScanSentences(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	var (
		start, end, rcount int
		head, nn           bool
	)
	head = true
	for p := 0; p < len(data); {
		r, size := utf8.DecodeRune(data[p:])
		if s.SkipWhiteSpace && unicode.IsSpace(r) {
			p += size
			if head {
				start, end = p, p
			} else if s.DoubleLineFeedSplit && r == '\n' {
				if nn {
					return p, data[start:end], nil
				}
				nn = true
			}
			continue
		}
		head, nn = false, false // clear flags
		if end != p {
			for i := 0; i < size; i++ {
				data[end+i] = data[p+i]
			}
		}
		p += size
		end += size
		rcount++
		if !s.isDelim(r) && rcount < s.MaxRuneLen {
			continue
		}
		// split
		nn = false
		for p < len(data) {
			r, size := utf8.DecodeRune(data[p:])
			if s.SkipWhiteSpace && unicode.IsSpace(r) {
				p += size
				if s.DoubleLineFeedSplit && r == '\n' {
					if nn {
						return p, data[start:end], nil
					}
					nn = true
				}
			} else if s.isDelim(r) || s.isFollower(r) {
				if end != p {
					for i := 0; i < size; i++ {
						data[end+i] = data[p+i]
					}
				}
				p += size
				end += size
			} else {
				break
			}
		}
		return p, data[start:end], nil
	}
	if !atEOF {
		// Request more data
		for i := end; i < len(data); i++ {
			data[i] = ' '
		}
		return start, nil, nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	return len(data), data[start:end], nil

}
