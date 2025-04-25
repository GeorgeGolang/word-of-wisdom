package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	serverAddr := os.Getenv("SERVER_ADDR")
	if serverAddr == "" {
		serverAddr = "localhost:8080"
	}
	fmt.Printf("Attempting to connect to server at %s\n", serverAddr)

	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		fmt.Println("Error connecting:", err)
		return
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)

	// 1. Получение и решение предварительного PoW
	conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	preChallenge, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading preliminary challenge:", err)
		return
	}
	preChallenge = strings.TrimSpace(preChallenge)
	fmt.Println("Received preliminary challenge:", preChallenge)

	if !strings.HasPrefix(preChallenge, "PRE:") {
		fmt.Println("Expected preliminary challenge with PRE: prefix")
		return
	}

	nonce, err := solvePoW(preChallenge[4:])
	if err != nil {
		fmt.Println("Error solving preliminary PoW:", err)
		return
	}

	conn.SetWriteDeadline(time.Now().Add(500 * time.Millisecond))
	fmt.Fprintf(conn, "%s\n", nonce)
	fmt.Println("Sent preliminary nonce:", nonce)

	// 2. Получение и решение основного PoW
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	challenge, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading main challenge:", err)
		return
	}
	challenge = strings.TrimSpace(challenge)
	fmt.Println("Received main challenge:", challenge)

	if strings.Contains(challenge, "Error") || strings.Contains(challenge, "Invalid") {
		fmt.Println("Server error:", challenge)
		return
	}

	nonce, err = solvePoW(challenge)
	if err != nil {
		fmt.Println("Error solving main PoW:", err)
		return
	}

	conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
	fmt.Fprintf(conn, "%s\n", nonce)
	fmt.Println("Sent main nonce:", nonce)

	// 3. Получение цитаты
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	response, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading response:", err)
		return
	}
	response = strings.TrimSpace(response)
	fmt.Println("Server response:", response)

	if strings.Contains(response, "Error") || strings.Contains(response, "Invalid") {
		fmt.Println("Failed to get quote:", response)
		return
	}
}

func solvePoW(challenge string) (string, error) {
	parts := strings.Split(challenge, ":")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid challenge format: %s", challenge)
	}
	value := parts[1]
	difficulty, err := strconv.Atoi(parts[2])
	if err != nil {
		return "", fmt.Errorf("invalid difficulty: %s", parts[2])
	}
	prefix := strings.Repeat("0", difficulty)

	for nonce := 0; ; nonce++ {
		input := value + strconv.Itoa(nonce)
		hash := sha256.Sum256([]byte(input))
		hashStr := hex.EncodeToString(hash[:])
		if strings.HasPrefix(hashStr, prefix) {
			return strconv.Itoa(nonce), nil
		}
	}
}
