package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"modak/mock"
	"modak/notification"
	"net/http"
	"os"
	"strings"
)

const (
	webhookPath = "/webhook/notification"
)

// This code snippet depicts a basic mockup webhook designed
// for utilization with the package notification service.
// It's important to note that this section of the code is not
// optimized for a production environment; rather, its purpose
// is for interaction and testing purposes.
func main() {
	port := os.Getenv("PORT")
	if port == "" {
		panic("PORT is required")
	}
	notifTTL := os.Getenv("CACHE_TTL")
	if notifTTL == "" {
		fmt.Println("CACHE_TTL not set/found")
	}

	fmt.Println("starting server at localhost", port)

	repo := mock.NewNotificationMemoryRepo()
	cache := mock.NewCacheMemoryRepo()
	gateway := mock.NewGateway()

	notifService := notification.NewNotificationService(
		gateway,
		repo,
		cache,
		notification.Config{CacheTTL: notifTTL},
	)

	http.HandleFunc(webhookPath, func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(w, fmt.Sprintf("method '%v', not allowed", req.Method), http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var notifReq Request
		if err := json.Unmarshal(body, &notifReq); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := notifReq.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = notifService.Send(notifReq.ToNotification())
		if errors.Is(err, notification.RateLimitError) {
			http.Error(w, err.Error(), http.StatusTooManyRequests)
			return
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	if err := http.ListenAndServe(fmt.Sprintf(":%v", port), nil); err != nil {
		log.Fatal(err)
	}

	fmt.Println("server closed")
}

type Request struct {
	Type    string `json:"type"`
	UserID  int    `json:"user_id"`
	Message string `json:"message"`
}

func (req *Request) Validate() error {
	var errs []error
	if strings.TrimSpace(req.Message) == "" {
		errs = append(errs, errors.New("message is required"))
	}

	if strings.TrimSpace(req.Type) == "" {
		errs = append(errs, errors.New("type is required"))
	}

	if req.UserID <= 0 {
		errs = append(errs, errors.New("user id is required"))
	}

	return errors.Join(errs...)
}

func (r Request) ToNotification() notification.Request {
	return notification.Request{
		Type:    r.Type,
		Message: r.Message,
		UserID:  r.UserID,
	}
}
