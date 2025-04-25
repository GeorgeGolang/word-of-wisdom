package repository

import (
	"fmt"
	"math/rand"
	"time"
	"word-of-wisdom/domain"
)

type QuoteGenerator struct {
	adjectives []string
	nouns      []string
	verbs      []string
	authors    []string
}

func NewQuoteGenerator() *QuoteGenerator {
	adjectives := make([]string, 100)
	nouns := make([]string, 100)
	verbs := make([]string, 100)
	authors := make([]string, 100)
	for i := 0; i < 100; i++ {
		adjectives[i] = fmt.Sprintf("adj%d", i)
		nouns[i] = fmt.Sprintf("noun%d", i)
		verbs[i] = fmt.Sprintf("verb%d", i)
		authors[i] = fmt.Sprintf("author%d", i)
	}
	return &QuoteGenerator{
		adjectives: adjectives,
		nouns:      nouns,
		verbs:      verbs,
		authors:    authors,
	}
}

func (g *QuoteGenerator) GenerateQuote(counter int64) (domain.Quote, error) {
	rand.Seed(time.Now().UnixNano())
	text := fmt.Sprintf("%s %s %s #%d",
		g.adjectives[rand.Intn(len(g.adjectives))],
		g.nouns[rand.Intn(len(g.nouns))],
		g.verbs[rand.Intn(len(g.verbs))],
		counter,
	)
	author := g.authors[rand.Intn(len(g.authors))]
	return domain.Quote{Text: text, Author: author}, nil
}
