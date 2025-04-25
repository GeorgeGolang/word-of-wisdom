package usecase

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"word-of-wisdom/domain"
	"word-of-wisdom/repository"
)

type QuoteService struct {
	repo       *repository.QuoteGenerator
	usedQuotes *sync.Map
	counter    int64
}

func NewQuoteService(repo *repository.QuoteGenerator) *QuoteService {
	return &QuoteService{
		repo:       repo,
		usedQuotes: &sync.Map{},
	}
}

func (s *QuoteService) GetQuote() (domain.Quote, error) {
	for i := 0; i < 100; i++ {
		counter := atomic.AddInt64(&s.counter, 1)
		quote, err := s.repo.GenerateQuote(counter)
		if err != nil {
			return domain.Quote{}, err
		}
		quoteStr := fmt.Sprintf("%s — %s", quote.Text, quote.Author)
		_, loaded := s.usedQuotes.LoadOrStore(quoteStr, true)
		if !loaded {
			return quote, nil
		}
	}
	return domain.Quote{}, errors.New("failed to generate unique quote")
}
