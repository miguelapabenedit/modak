// Package mock provides a simple in-memory dummy implementation of the notification service dependecies.
// It provides a simple in-memory cache system with key-value storage, console sender(gateway) implementation
// that writes console i/o and a basic storer repository that holds the notification data.
package mock

import (
	"encoding/json"
	"errors"
	"fmt"
	"modak/notification"
	"strings"
	"time"
)

type gateway struct {
}

func NewGateway() notification.Sender {
	return &gateway{}
}

func (g *gateway) Send(userID int, message string) error {
	_, err := fmt.Printf("sending message to user_id %v, msg:%s\n", userID, message)
	return err
}

type memoryRepo struct {
	notifications notification.Notifications
	types         notification.Types
}

func NewNotificationMemoryRepo() notification.Storer {
	repo := &memoryRepo{}
	repo.types = append(repo.types,
		notification.Type{ID: 1, Name: "STATUS", RateLimit: 120, RequestLimit: 2, IsEnable: true},
		notification.Type{ID: 2, Name: "NEWS", RateLimit: 86400, RequestLimit: 1, IsEnable: true},
		notification.Type{ID: 2, Name: "MARKETING", RateLimit: 10800, RequestLimit: 3, IsEnable: true},
	)

	return repo
}

func (r *memoryRepo) GetUserNotificationSinceDate(
	userID int,
	ntype string,
	startDate time.Time,
) (notification.Notifications, error) {
	var data notification.Notifications
	if len(r.notifications) == 0 {
		return notification.Notifications{}, nil
	}

	for i := len(r.notifications) - 1; i >= 0; i-- {
		if strings.EqualFold(r.notifications[i].Type.Name, ntype) &&
			r.notifications[i].UserID == userID &&
			r.notifications[i].Date.After(startDate) {
			data = append(data, r.notifications[i])
		}
	}

	return data, nil
}

func (r *memoryRepo) GetTypes() (notification.Types, error) {
	types := notification.Types{}
	for _, t := range r.types {
		if t.IsEnable {
			types = append(types, t)
		}
	}

	return types, nil
}

func (r *memoryRepo) Save(notifType notification.Type, userID int) error {
	notif := notification.NewNotification(userID, notifType)
	notif.ID = len(r.notifications) + 1
	notif.Date = time.Now()
	r.notifications = append(r.notifications, notif)
	return nil
}

type memoryCache struct {
	Items map[string][]byte
}

func NewCacheMemoryRepo() notification.Cacher {
	items := make(map[string][]byte)
	return &memoryCache{items}
}

func (mc *memoryCache) Get(key string, obj interface{}) error {
	data, ok := mc.Items[key]
	if !ok || len(data) == 0 {
		return errors.New("key not found")
	}

	return json.Unmarshal(data, &obj)
}

// Set stores a new object within a given key in memory. ttl sets the reminaing time
// of the key before timeout.
//
// **For simplicity the ttl param has no effect on this implementation**
func (mc *memoryCache) Set(key string, value interface{}, _ time.Duration) error {
	bytes, err := json.Marshal(value)
	if err != nil {
		return err
	}
	mc.Items[key] = bytes
	return nil
}
