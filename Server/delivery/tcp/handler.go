package tcp

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
	"word-of-wisdom/usecase"
)

type ConnectionHandlerImpl struct {
	powService   *usecase.PoWService
	quoteService *usecase.QuoteService
}

func NewConnectionHandler(powService *usecase.PoWService, quoteService *usecase.QuoteService) *ConnectionHandlerImpl {
	return &ConnectionHandlerImpl{
		powService:   powService,
		quoteService: quoteService,
	}
}

func (h *ConnectionHandlerImpl) HandleConnection(conn net.Conn) {
	log.Printf("Connection from %s", conn.RemoteAddr())
	defer conn.Close()

	// Предварительное PoW (difficulty=3) (0.5 секунды)
	preChallenge := h.powService.GenerateChallenge(3)
	fmt.Fprintf(conn, "PRE:%s:%s:%d\n", preChallenge.ID, preChallenge.Value, preChallenge.Difficulty)

	conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	reader := bufio.NewReader(conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		fmt.Fprintf(conn, "Error reading preliminary response: %v\n")
		h.powService.MarkSuspicious(conn.RemoteAddr().(*net.TCPAddr).IP.String())
		return
	}
	preNonce := strings.TrimSpace(response)
	if !h.powService.VerifyChallenge(preChallenge, preNonce) {
		fmt.Fprintf(conn, "Invalid preliminary PoW\n")
		h.powService.MarkSuspicious(conn.RemoteAddr().(*net.TCPAddr).IP.String())
		return
	}

	// Основное PoW (2 секунды)
	challenge := h.powService.GenerateChallenge(4)
	fmt.Fprintf(conn, "%s:%s:%d\n", challenge.ID, challenge.Value, challenge.Difficulty)

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	response, err = reader.ReadString('\n')
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			fmt.Fprintf(conn, "Timeout: no response received\n")
		} else {
			fmt.Fprintf(conn, "Error reading response: %v\n")
		}
		h.powService.MarkSuspicious(conn.RemoteAddr().(*net.TCPAddr).IP.String())
		return
	}
	nonce := strings.TrimSpace(response)

	// Проверка PoW
	if h.powService.VerifyChallenge(challenge, nonce) {
		quote, err := h.quoteService.GetQuote()
		if err != nil {
			fmt.Fprintf(conn, "Error getting quote: %v\n", err)
			return
		}
		fmt.Fprintf(conn, "%s — %s\n", quote.Text, quote.Author)
	} else {
		fmt.Fprintf(conn, "Invalid PoW\n")
		h.powService.MarkSuspicious(conn.RemoteAddr().(*net.TCPAddr).IP.String())
	}
}
