//package main
//
//import (
//	"POWDDOS/delivery/tcp"
//	"POWDDOS/repository"
//	"POWDDOS/usecase"
//)
//
//func main() {
//	// Инициализация репозитория
//	quoteRepo := repository.NewInMemoryQuoteRepository()
//
//	// Инициализация сервисов
//	powService := usecase.NewPoWService()
//	quoteService := usecase.NewQuoteService(quoteRepo)
//
//	// Настройка TCP-сервера
//	server := tcp.NewServer(":8080", powService, quoteService)
//	if err := server.Start(); err != nil {
//		panic(err)
//	}
//}

package main

import (
	"log"
	"os"
	"word-of-wisdom/delivery/tcp"
	"word-of-wisdom/repository"
	"word-of-wisdom/usecase"
)

func main() {
	addr := os.Getenv("PORT")
	if addr == "" {
		addr = ":8080"
	}

	quoteRepo := repository.NewQuoteGenerator()
	quoteService := usecase.NewQuoteService(quoteRepo)
	powService := usecase.NewPoWService()

	server := tcp.NewServer(addr, powService, quoteService)
	if err := server.Start(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
