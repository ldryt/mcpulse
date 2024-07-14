package announcer

import (
	"encoding/json"
	"io"
	"log"
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

func writeAnnouncements(w io.Writer, tick time.Duration) {
	ticker := time.NewTicker(tick * time.Second)
	defer ticker.Stop()

	encoder := json.NewEncoder(w)
	for range ticker.C {
		err := encoder.Encode(container)
		if err != nil {
			log.Printf("error sending announcement: %v", err)
			break
		}
	}
}

func readUpdates(r io.Reader) {
	decoder := json.NewDecoder(r)
	for {
		var msg _StartRequest
		err := decoder.Decode(&msg)
		if err == io.EOF {
			continue
		} else {
			log.Println("error decoding update message:", err)
			break
		}
	}
}
