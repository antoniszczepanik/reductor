package main

import (
	"bytes"
	"sync"
)

func getLongestMatchPosAndLen(text, pattern []byte, minMatchLen byte) (int, byte) {
	if len(pattern) < int(minMatchLen) {
		return 0, 0
	}
	var (
		matchLen, maxSoFar, length byte
		position                   int
	)
	// Heuristic: get indexes at which at least minMatchLen of pattern matches.
	minMatchStarts := getMatchIndices(text, pattern[:minMatchLen])
	for _, matchStart := range minMatchStarts {
		matchLen = getMatchLen(text[matchStart:], pattern)
		if matchLen >= minMatchLen && matchLen > maxSoFar {
			position = matchStart
			length, maxSoFar = matchLen, matchLen
		}
	}
	return position, length
}

// matchIndices stores indices of matches found protected by mu.
type matchIndices struct {
	// Indexes of starting position of a match.
	matches []int
	mu      sync.Mutex
}

func (mi *matchIndices) addMatches(ms []int) {
	mi.mu.Lock()
	mi.matches = append(mi.matches, ms...)
	mi.mu.Unlock()
}

// getMatchIndices will return a slice of indexes at which pattern begins.
func getMatchIndices(text, pattern []byte) []int {
	const chunkSize = 1024
	mi := &matchIndices{}
	splits := getSplits(text, chunkSize)
	var wg sync.WaitGroup
	for _, s := range splits {
		end := min(s+chunkSize+len(pattern)-1, len(text))
		wg.Add(1)
		go getStarts(text[s:end], pattern, s, mi, &wg)
	}
	wg.Wait()
	return mi.matches
}

func getStarts(text, pattern []byte, offset int, mi *matchIndices, wg *sync.WaitGroup) {
	defer wg.Done()
	if len(text) == 0 || len(text) < len(pattern) {
		return
	}
	// If pattern is empty, then we have a match everywhere.
	if len(pattern) == 0 {
		starts := make([]int, len(text))
		for i := range text {
			starts[i] = offset + i
		}
		mi.addMatches(starts)
		return
	}
	starts := make([]int, 0)
	for i := range text[:len(text)-len(pattern)+1] {
		// First compare a single byte.
		if text[i] == pattern[0] {
			// If single byte matches, try to compare all bytes.
			if bytes.Equal(text[i:i+len(pattern)], pattern) {
				starts = append(starts, offset+i)
			}
		}
	}
	mi.addMatches(starts)
	return
}

func getSplits(text []byte, chunkSize int) []int {
	splits := make([]int, 0)
	for i := 0; i < len(text); i += chunkSize {
		splits = append(splits, i)
	}
	return splits
}

// getMatchLen returns a length of a longest match between two sequences.
func getMatchLen(a, b []byte) byte {
	var matchLen byte
	maxMatchLen := min(min(len(a), len(b)), 255)
	for i := 0; i < maxMatchLen; i++ {
		if a[i] == b[i] {
			matchLen += 1
		} else {
			break
		}
	}
	return matchLen
}
