package notification

import (
	"errors"
	"fmt"
	"time"
)

const notificationTypeKey = "notification_types"

var RateLimitError = errors.New("notification rate limit reached")

// Storer is implemented by any value that has the base methods required
// in the implementation of notification service storage unit.
type Storer interface {
	// GetUserNotificationSinceDate retrieves all user notifications by type between the startDate and the
	// and time.Now().
	GetUserNotificationSinceDate(userID int, ntype string, startDate time.Time) (Notifications, error)
	// GetTypes returns all the available notification types.
	GetTypes() (Types, error)
	// Save inserts the last notification sent to an user by userID and type of notification.
	Save(notifType Type, userID int) error
}

// Cacher is implemented by any value that has Get and Set method. This
// interface provides the base cache functionalities required.
type Cacher interface {
	Get(key string, i interface{}) error
	Set(key string, value interface{}, ttl time.Duration) error
}

// Sender is implemented by any value that has a Send method. The
// implementation controls how/where the notification message is sent
// as a gateway.
type Sender interface {
	Send(userID int, message string) error
}

// Config provides configurations to the service.
//
// CacheTTL - This function updates the valid time of a key before
// the timeout is reached in the memory cache system. Its a string
// that use the time.ParseDuration native function. For more information
// of the format please review https://pkg.go.dev/time#ParseDuration
type Config struct {
	CacheTTL string
}

type NotificationService struct {
	gateway    Sender
	repository Storer
	cache      Cacher
	ttl        time.Duration
}

// NotificationService offers a comprehensive set of functionalities
// for efficiently managing and processing upcoming notification
// requests intended to be delivered. Accesing the struct directly is
// not safe,use this function instead.
func NewNotificationService(
	gateway Sender,
	repository Storer,
	cache Cacher,
	config Config,
) *NotificationService {
	ttl, err := time.ParseDuration(config.CacheTTL)
	if err != nil {
		panic("required cache ttl invalid or not found")
	}

	return &NotificationService{
		gateway,
		repository,
		cache,
		ttl,
	}
}

// Send process the upcoming notification request by performing some validations like rate limit. In
// case the rate limit is reached a sentinel RateLimitError is provided in return.
func (ns *NotificationService) Send(req Request) error {
	if err := req.Validate(); err != nil {
		return err
	}

	notifType, err := ns.getValidatedTypeByName(req.Type)
	if err != nil {
		return err
	}

	if err := ns.validateRateLimit(notifType, req.UserID); err != nil {
		return err
	}

	if err := ns.gateway.Send(req.UserID, req.Message); err != nil {
		return err
	}

	return ns.repository.Save(notifType, req.UserID)
}

func (ns *NotificationService) getValidatedTypeByName(name string) (Type, error) {
	var (
		types     Types
		notifType Type
	)
	if err := ns.cache.Get(notificationTypeKey, &types); err == nil {
		notifType = types.GetByName(name)
	} else {
		types, err = ns.repository.GetTypes()
		if err != nil {
			return Type{}, err
		}

		if err := ns.cache.Set(notificationTypeKey, &types, ns.ttl); err != nil {
			return Type{}, err
		}

		notifType = types.GetByName(name)
	}

	if !notifType.Exists() {
		return Type{}, fmt.Errorf("notification type '%s' not found or disabled", name)
	}

	return notifType, nil
}

func (ns *NotificationService) validateRateLimit(notifType Type, userID int) error {
	rateLimitDuration := (time.Second * time.Duration(notifType.RateLimit))
	startDate := time.Now().Add(-rateLimitDuration)
	notifications, err := ns.repository.GetUserNotificationSinceDate(userID, notifType.Name, startDate)
	if err != nil {
		return err
	}

	if notifications.HasAny() {
		if len(notifications) >= notifType.RequestLimit {
			return RateLimitError
		}
	}

	return nil
}
