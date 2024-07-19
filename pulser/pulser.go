package pulser

import (
	"encoding/json"
	"io"
	"log"
	"sync"
	"time"
)

type _StartRequest struct {
	Time int64  `json:"time"`
	UUID string `json:"uuid"`
}

type _StartRequestsContainer struct {
	StartRequests []_StartRequest `json:"startRequests"`
}

var container _StartRequestsContainer

func GetStartRequestsCount() int {
	return len(container.StartRequests)
}

func AddStartRequest(UUID string) {
	newStartRequest := _StartRequest{UUID: UUID, Time: time.Now().UnixMilli()}
	container.StartRequests = append(container.StartRequests, newStartRequest)
}

func pulse(w io.Writer, tick time.Duration, wg *sync.WaitGroup) {
	defer wg.Done()

	ticker := time.NewTicker(tick * time.Second)
	defer ticker.Stop()

	encoder := json.NewEncoder(w)
	for range ticker.C {
		err := encoder.Encode(container)
		if err != nil {
			log.Printf("error sending announcement: %v", err)
			return
		}
	}
}

func readUpdates(r io.Reader, wg *sync.WaitGroup) {
	defer wg.Done()

	decoder := json.NewDecoder(r)
	for {
		var msg _StartRequest
		err := decoder.Decode(&msg)
		if err == io.EOF {
			continue
		} else {
			log.Println("error decoding update message:", err)
			return
		}
	}
}
