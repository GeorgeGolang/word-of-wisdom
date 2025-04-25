//package repository
//
//import "POWDDOS/domain"
//
//type InMemoryQuoteRepository struct {
//	quotes []domain.Quote
//}
//
//func NewInMemoryQuoteRepository() *InMemoryQuoteRepository {
//	return &InMemoryQuoteRepository{
//		quotes: []domain.Quote{
//			{Text: "Stay hungry, stay foolish.", Author: "Steve Jobs"},
//			{Text: "Stay hungry, stay foolish2", Author: "Steve Jobs"},
//			{Text: "Stay hungry, stay foolish3", Author: "Steve Jobs"},
//			{Text: "Stay hungry, stay foolish4", Author: "Steve Jobs"},
//			{Text: "Stay hungry, stay foolish5", Author: "Steve Jobs"},
//			{Text: "Stay hungry, stay foolish6", Author: "Steve Jobs"},
//			{Text: "Stay hungry, stay foolish7", Author: "Steve Jobs"},
//			{Text: "Stay hungry, stay foolish8", Author: "Steve Jobs"},
//			{Text: "Stay hungry, stay foolish9", Author: "Steve Jobs"},
//			{Text: "Stay hungry, stay foolish10", Author: "Steve Jobs"},
//
//			{Text: "Life is what happens when you're busy making other plans.", Author: "John Lennon"},
//		},
//	}
//}
//
//func (r *InMemoryQuoteRepository) GetAll() ([]domain.Quote, error) {
//	return r.quotes, nil
//}

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
