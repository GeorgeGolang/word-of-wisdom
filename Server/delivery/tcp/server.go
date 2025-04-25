//package tcp
//
//import (
//	"POWDDOS/usecase"
//	"fmt"
//	"log"
//	"net"
//)
//
//// ConnectionHandler интерфейс для обработки TCP-соединений
//type ConnectionHandler interface {
//	HandleConnection(conn net.Conn)
//}
//
//type Server struct {
//	addr    string
//	handler ConnectionHandler
//}
//
//func NewServer(addr string, powService *usecase.PoWService, quoteService *usecase.QuoteService) *Server {
//	return &Server{
//		addr:    addr,
//		handler: NewConnectionHandler(powService, quoteService),
//	}
//}
//
//func (s *Server) Start() error {
//	listener, err := net.Listen("tcp", s.addr)
//	if err != nil {
//		return fmt.Errorf("failed to start server: %v", err)
//	}
//	log.Printf("TCP server running on %s", s.addr)
//	defer listener.Close()
//
//	for {
//		conn, err := listener.Accept()
//		if err != nil {
//			log.Printf("Error accepting connection: %v", err)
//			continue
//		}
//		go s.handler.HandleConnection(conn)
//	}
//}

package tcp

import (
	"fmt"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"
	"word-of-wisdom/usecase"
)

type Server struct {
	addr      string
	handler   ConnectionHandler
	ipCount   *sync.Map
	connChan  chan net.Conn
	connCount int64
	conns     *sync.Map
}

type ConnectionHandler interface {
	HandleConnection(conn net.Conn)
}

func NewServer(addr string, powService, quoteService interface{}) *Server {
	pow, ok := powService.(*usecase.PoWService)
	if !ok {
		log.Fatalf("powService must be *usecase.PoWService, got %T", powService)
	}
	quote, ok := quoteService.(*usecase.QuoteService)
	if !ok {
		log.Fatalf("quoteService must be *usecase.QuoteService, got %T", quoteService)
	}
	return &Server{
		addr:     addr,
		handler:  NewConnectionHandler(pow, quote),
		ipCount:  &sync.Map{},
		connChan: make(chan net.Conn, 2000),
		conns:    &sync.Map{},
	}
}

func (s *Server) Start() error {
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	log.Printf("TCP server running on %s", s.addr)
	sem := make(chan struct{}, 8000)
	go s.processConnChan(sem)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}
		ip := conn.RemoteAddr().(*net.TCPAddr).IP.String()
		if s.handler.(*ConnectionHandlerImpl).powService.IsSuspicious(ip) {
			conn.Close()
			continue
		}
		count, _ := s.ipCount.LoadOrStore(ip, int64(0))
		newCount := count.(int64) + 1
		if newCount > 20 {
			conn.Close()
			continue
		}
		s.ipCount.Store(ip, newCount)
		select {
		case sem <- struct{}{}:
			atomic.AddInt64(&s.connCount, 1)
			connID := fmt.Sprintf("%d", time.Now().UnixNano())
			s.conns.Store(connID, conn)
			go func() {
				defer func() {
					<-sem
					atomic.AddInt64(&s.connCount, -1)
					s.conns.Delete(connID)
					if count, ok := s.ipCount.Load(ip); ok {
						s.ipCount.Store(ip, count.(int64)-1)
					}
					s.handler.(*ConnectionHandlerImpl).powService.DecrementConns()
				}()
				s.handler.(*ConnectionHandlerImpl).powService.IncrementConns()
				s.handler.HandleConnection(conn)
			}()
		case s.connChan <- conn:
			// В очередь
		default:
			conn.Close()
		}
	}
}

func (s *Server) processConnChan(sem chan struct{}) {
	for conn := range s.connChan {
		ip := conn.RemoteAddr().(*net.TCPAddr).IP.String()
		if s.handler.(*ConnectionHandlerImpl).powService.IsSuspicious(ip) {
			conn.Close()
			continue
		}
		count, _ := s.ipCount.LoadOrStore(ip, int64(0))
		newCount := count.(int64) + 1
		if newCount > 20 {
			conn.Close()
			continue
		}
		s.ipCount.Store(ip, newCount)
		select {
		case sem <- struct{}{}:
			if atomic.LoadInt64(&s.connCount) >= 8000 {
				var oldestID string
				s.conns.Range(func(key, value interface{}) bool {
					oldestID = key.(string)
					return false
				})
				if oldestID != "" {
					if oldConn, ok := s.conns.LoadAndDelete(oldestID); ok {
						oldConn.(net.Conn).Close()
						log.Printf("Closed old connection %s to free slot", oldestID)
					}
				}
			}
			atomic.AddInt64(&s.connCount, 1)
			connID := fmt.Sprintf("%d", time.Now().UnixNano())
			s.conns.Store(connID, conn)
			go func() {
				defer func() {
					<-sem
					atomic.AddInt64(&s.connCount, -1)
					s.conns.Delete(connID)
					if count, ok := s.ipCount.Load(ip); ok {
						s.ipCount.Store(ip, count.(int64)-1)
					}
					s.handler.(*ConnectionHandlerImpl).powService.DecrementConns()
				}()
				s.handler.(*ConnectionHandlerImpl).powService.IncrementConns()
				s.handler.HandleConnection(conn)
			}()
		case <-time.After(500 * time.Millisecond):
			conn.Close()
		}
	}
}
