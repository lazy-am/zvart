package server

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/lazy-am/zvart/internal/core/contact"
	"github.com/lazy-am/zvart/internal/storage"
	"github.com/lazy-am/zvart/pkg/service"
)

const (
	secBetweenConnectionAttempts = 20
)

type Server struct {
	listener   net.Listener
	errChan    chan error
	ctx        context.Context
	cancelFunc context.CancelFunc
	storage    storage.Storage
	socksPort  uint16
}

func Init(l net.Listener, db storage.Storage, socksPort uint16) (*Server, error) {
	s := Server{errChan: make(chan error, 1), listener: l, storage: db, socksPort: socksPort}
	s.ctx, s.cancelFunc = context.WithCancel(context.Background())
	s.resetWork()

	go func() { s.errChan <- http.Serve(s.listener, s.initOnionMux()) }()
	go s.ioutgoingConnections()
	service.AddService(&s)
	return &s, nil
}

func (s *Server) Close() {
	s.cancelFunc()
	s.listener.Close()
}

func (s *Server) ioutgoingConnections() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.checkContacts()
		}
	}
}

func (s *Server) checkContacts() {
	cl, err := contact.LoadList(s.storage)
	if err != nil {
		return
	}
	for _, c := range cl {
		if c.InWork {
			continue
		}
		if c.CreatedByMe &&
			!c.PubKeySended &&
			(int(time.Since(c.LastTryTime).Seconds()) > secBetweenConnectionAttempts) {
			go s.registSender(c)
		}
		if c.PubKey != nil &&
			c.FirstUnsentMessageId != nil &&
			(int(time.Since(c.LastTryTime).Seconds()) > secBetweenConnectionAttempts) {
			go s.sendMessages(c)
		}
	}
}

// Reset the operation flags on restart
func (s *Server) resetWork() {
	cl, err := contact.LoadList(s.storage)
	if err != nil {
		return
	}
	for _, c := range cl {
		if c.InWork {
			c.InWork = false
			c.Save(s.storage)
		}
	}
}

func (s *Server) initOnionMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/"+registrationAddress, s.registHandler)
	mux.HandleFunc("/"+contregistAddress, s.regist2Handler)
	mux.HandleFunc("/"+textmessageAddress, s.textmesHandler)
	mux.HandleFunc("/"+seskeyAddress, s.sesKeyHandler)
	//mux.HandleFunc("/", s.testJSON)
	return mux
}

// func (s *Server) testJSON(rw http.ResponseWriter, req *http.Request) {
// 	var data map[string]string
// 	decoder := json.NewDecoder(req.Body)
// 	err := decoder.Decode(&data)
// 	if err != nil {
// 		http.Error(rw, "Error decoding JSON", http.StatusBadRequest)
// 		return
// 	}
// 	fmt.Print(data)

// 	// Use the data as needed
// 	// ...
// }
