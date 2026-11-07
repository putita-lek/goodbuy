package service

import (
	"strings"
	"unicode"
)

// Words that describe how a shopper phrases an item rather than what the item is.
var stopWords = map[string]bool{
	"a": true, "an": true, "the": true, "of": true, "in": true, "for": true,
	"with": true, "and": true, "some": true, "my": true, "to": true, "by": true,
}

// Tokenize lower-cases text, splits on non-alphanumerics, drops stop words and
// strips a simple plural "s" so "apples" matches "apple".
func Tokenize(text string) []string {
	fields := strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
	seen := map[string]bool{}
	var tokens []string
	for _, f := range fields {
		if stopWords[f] {
			continue
		}
		f = stem(f)
		if f == "" || seen[f] {
			continue
		}
		seen[f] = true
		tokens = append(tokens, f)
	}
	return tokens
}

func stem(w string) string {
	if len(w) > 3 && strings.HasSuffix(w, "s") && !strings.HasSuffix(w, "ss") {
		return w[:len(w)-1]
	}
	return w
}

// MatchScore is the share of query tokens found in the product's name or keywords (0..1).
func MatchScore(queryTokens []string, productName, keywords string) float64 {
	if len(queryTokens) == 0 {
		return 0
	}
	productTokens := map[string]bool{}
	for _, t := range Tokenize(productName + " " + keywords) {
		productTokens[t] = true
	}
	hits := 0
	for _, q := range queryTokens {
		if productTokens[q] {
			hits++
		}
	}
	return float64(hits) / float64(len(queryTokens))
}
