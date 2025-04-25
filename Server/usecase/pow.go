package usecase

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"word-of-wisdom/domain"
)

type PoWService struct {
	challenges       *sync.Map // Активные задачи
	solvedChallenges *sync.Map // Решённые задачи
	activeConns      int64
	suspiciousIPs    *sync.Map // Подозрительные IP
}

func NewPoWService() *PoWService {
	return &PoWService{
		challenges:       &sync.Map{},
		solvedChallenges: &sync.Map{},
		suspiciousIPs:    &sync.Map{},
	}
}

func (s *PoWService) GenerateChallenge(difficulty int) domain.Challenge {
	idBytes := make([]byte, 8)
	rand.Read(idBytes)
	valueBytes := make([]byte, 8)
	rand.Read(valueBytes)
	adjustedDifficulty := difficulty
	if atomic.LoadInt64(&s.activeConns) > 8000 {
		adjustedDifficulty = 6 // ~3.4 с
	}
	challenge := domain.Challenge{
		ID:         fmt.Sprintf("%x", idBytes),
		Value:      fmt.Sprintf("%x", valueBytes),
		Difficulty: adjustedDifficulty,
	}
	s.challenges.Store(challenge.ID, challenge)
	go func() {
		time.Sleep(10 * time.Second)
		s.challenges.Delete(challenge.ID)
	}()
	return challenge
}

func (s *PoWService) VerifyChallenge(challenge domain.Challenge, nonce string) bool {
	if _, exists := s.challenges.Load(challenge.ID); !exists {
		return false
	}

	// Проверка, не была ли задача уже решена
	challengeKey := fmt.Sprintf("%s:%s:%d:%s", challenge.ID, challenge.Value, challenge.Difficulty, nonce)
	if _, exists := s.solvedChallenges.Load(challengeKey); exists {
		return false // Задача уже решена
	}
	hash := sha256.Sum256([]byte(challenge.Value + nonce))
	hashStr := hex.EncodeToString(hash[:])
	if strings.HasPrefix(hashStr, strings.Repeat("0", challenge.Difficulty)) {
		s.challenges.Delete(challenge.ID)
		s.solvedChallenges.Store(challengeKey, true)
		go func() {
			time.Sleep(60 * time.Second)
			s.solvedChallenges.Delete(challengeKey)
		}()
		return true
	}
	return false
}

func (s *PoWService) IncrementConns() {
	atomic.AddInt64(&s.activeConns, 1)
}

func (s *PoWService) DecrementConns() {
	atomic.AddInt64(&s.activeConns, -1)
}

func (s *PoWService) MarkSuspicious(ip string) {
	count, _ := s.suspiciousIPs.LoadOrStore(ip, int64(0))
	newCount := count.(int64) + 1
	s.suspiciousIPs.Store(ip, newCount)
	if newCount > 5 {
		go func() {
			time.Sleep(30 * time.Second)
			s.suspiciousIPs.Delete(ip)
		}()
	}
}

func (s *PoWService) IsSuspicious(ip string) bool {
	count, exists := s.suspiciousIPs.Load(ip)
	return exists && count.(int64) > 5
}
