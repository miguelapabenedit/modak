// Package test  furnishes structures that serve as mock implementations
// of the interface contracts. These mock structures are intended to
// facilitate testing and simulation scenarios, allowing for more
// controlled and predictable testing environments.
package test

import (
	"modak/notification"
	"time"
)

type MockCacher struct {
	Getter func(key string, i interface{}) error
	Setter func(key string, value interface{}, ttl time.Duration) error
}

func (m MockCacher) Get(key string, i interface{}) error {
	return m.Getter(key, i)
}
func (m MockCacher) Set(key string, value interface{}, ttl time.Duration) error {
	return m.Setter(key, value, ttl)
}

type MockStorer struct {
	GetterUserNotificationsFromDate func(
		userID int,
		ntype string,
		startDate time.Time,
	) (notification.Notifications, error)
	GetterTypes func() (notification.Types, error)
	Setter      func(notifType notification.Type, userID int) error
}

func (m MockStorer) GetUserNotificationSinceDate(
	userID int,
	ntype string,
	startDate time.Time,
) (notification.Notifications, error) {
	return m.GetterUserNotificationsFromDate(userID, ntype, startDate)
}
func (m MockStorer) GetTypes() (notification.Types, error) {
	return m.GetterTypes()
}
func (m MockStorer) Save(notifType notification.Type, userID int) error {
	return m.Setter(notifType, userID)
}

type MockSender struct {
	Sender func(userID int, message string) error
}

func (m MockSender) Send(userID int, message string) error {
	return m.Sender(userID, message)
}
