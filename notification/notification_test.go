package notification_test

import (
	"errors"
	"modak/notification"
	"modak/notification/test"
	"testing"
	"time"

	"gopkg.in/stretchr/testify.v1/assert"
)

func TestNew_WhenConfigEmpty_ThenPanic(t *testing.T) {
	expected := "required cache ttl invalid or not found"
	defer func() {
		r := recover()
		assert.NotNil(t, r)
		assert.Equal(t, expected, r)
	}()

	notification.NewNotificationService(nil, nil, nil, notification.Config{})
}

func TestNew_WhenConfigPresent_ThenNotNilSuccess(t *testing.T) {
	config := notification.Config{CacheTTL: "1h"}

	notifService := notification.NewNotificationService(nil, nil, nil, config)

	assert.NotNil(t, notifService)
}

func TestSend_WhenEmptyUserID_ThenFails(t *testing.T) {
	var (
		notifService = buildDefaultService(nil, nil, nil)
		req          = notification.Request{Type: "type_test", Message: "message_test"}
	)

	err := notifService.Send(req)

	assert.NotNil(t, err)
	assert.EqualError(t, err, "user id is required")
}

func TestSend_WhenEmptyMessage_ThenFails(t *testing.T) {
	var (
		notifService = buildDefaultService(nil, nil, nil)
		req          = notification.Request{Type: "type_test", Message: "", UserID: 1}
	)

	err := notifService.Send(req)

	assert.NotNil(t, err)
	assert.EqualError(t, err, "message is required")
}

func TestSend_WhenEmptyType_ThenFails(t *testing.T) {
	var (
		notifService = buildDefaultService(nil, nil, nil)
		req          = notification.Request{Type: "    ", Message: "message_test", UserID: 1}
	)

	err := notifService.Send(req)

	assert.NotNil(t, err)
	assert.EqualError(t, err, "type is required")
}

func TestSend_WhenEmptyRequest_ThenFails(t *testing.T) {
	var (
		notifService = buildDefaultService(nil, nil, nil)
		req          = notification.Request{}
	)

	err := notifService.Send(req)

	assert.NotNil(t, err)
	assert.EqualError(t, err, "message is required\ntype is required\nuser id is required")
}

func TestSend_WhenInvalidTypeError_ThenFails(t *testing.T) {
	var (
		cacherMock   = test.MockCacher{Getter: func(key string, i interface{}) error { return nil }}
		notifService = buildDefaultService(nil, nil, cacherMock)
		req          = notification.Request{Type: "type_test", Message: "message_test", UserID: 1}
	)

	err := notifService.Send(req)

	assert.NotNil(t, err)
	assert.EqualError(t, err, "notification type 'type_test' not found or disabled")
}

func TestSend_WhenGetTypesError_ThenFails(t *testing.T) {
	var (
		expErr       = "storer_test_err"
		storerMock   = test.MockStorer{GetterTypes: func() (notification.Types, error) { return nil, errors.New(expErr) }}
		notifService = buildDefaultService(nil, storerMock, nil)
		req          = notification.Request{Type: "type_test", Message: "message_test", UserID: 1}
	)

	err := notifService.Send(req)

	assert.NotNil(t, err)
	assert.EqualError(t, err, expErr)
}

func TestSend_WhenSetCacheError_ThenFails(t *testing.T) {
	var (
		expErr    = "set_cache_err"
		cacheMock = test.MockCacher{
			Getter: func(key string, i interface{}) error { return errors.New("expected_err") },
			Setter: func(key string, value interface{}, ttl time.Duration) error { return errors.New(expErr) },
		}
		notifService = buildDefaultService(nil, nil, cacheMock)
		req          = notification.Request{Type: "type_test", Message: "message_test", UserID: 1}
	)

	err := notifService.Send(req)

	assert.NotNil(t, err)
	assert.EqualError(t, err, expErr)
}

func TestSend_WhenGetUserNotificationsFromDateError_ThenFails(t *testing.T) {
	expErr := "set_cache_err"
	req := notification.Request{Type: "type_test", Message: "message_test", UserID: 1}
	storerMock := buildDefaultStorer()
	storerMock.GetterUserNotificationsFromDate = func(
		userID int,
		ntype string,
		startDate time.Time,
	) (notification.Notifications, error) {
		return nil, errors.New(expErr)
	}
	notifService := buildDefaultService(nil, storerMock, nil)

	err := notifService.Send(req)

	assert.NotNil(t, err)
	assert.EqualError(t, err, expErr)
}

func TestSend_WhenRequestRateLimitReached_ThenFails(t *testing.T) {
	req := notification.Request{Type: "type_test", Message: "message_test", UserID: 1}
	storerMock := buildDefaultStorer()
	storerMock.GetterUserNotificationsFromDate = func(
		userID int,
		ntype string,
		startDate time.Time,
	) (notification.Notifications, error) {
		return notification.Notifications{{ID: 1, UserID: req.UserID, Type: notification.Type{Name: req.Type}}}, nil
	}
	notifService := buildDefaultService(nil, storerMock, nil)

	err := notifService.Send(req)

	assert.NotNil(t, err)
	assert.EqualError(t, err, notification.RateLimitError.Error())
}

func TestSend_WhenSendError_ThenFails(t *testing.T) {
	var (
		expErr       = "sender_test_err"
		req          = notification.Request{Type: "type_test", Message: "message_test", UserID: 1}
		senderMock   = test.MockSender{Sender: func(userID int, message string) error { return errors.New(expErr) }}
		notifService = buildDefaultService(senderMock, nil, nil)
	)

	err := notifService.Send(req)

	assert.NotNil(t, err)
	assert.EqualError(t, err, expErr)
}

func TestSend_WhenAllOK_ThenSuccess(t *testing.T) {
	var (
		req          = notification.Request{Type: "type_test", Message: "message_test", UserID: 1}
		senderMock   = test.MockSender{Sender: func(userID int, message string) error { return nil }}
		notifService = buildDefaultService(senderMock, nil, nil)
	)

	err := notifService.Send(req)

	assert.Nil(t, err)
	assert.NoError(t, err)
}

func buildDefaultService(
	sender notification.Sender,
	storer notification.Storer,
	cacher notification.Cacher,
) *notification.NotificationService {
	if storer == nil {
		storer = buildDefaultStorer()
	}

	if cacher == nil {
		cacher = buildDefaultCacher()
	}

	return notification.NewNotificationService(sender, storer, cacher, notification.Config{CacheTTL: "1h"})
}

func buildDefaultCacher() test.MockCacher {
	return test.MockCacher{
		Getter: func(key string, i interface{}) error { return errors.New("intended_error_test") },
		Setter: func(key string, value interface{}, ttl time.Duration) error { return nil },
	}
}

func buildDefaultStorer() test.MockStorer {
	return test.MockStorer{
		GetterUserNotificationsFromDate: func(
			userID int,
			ntype string,
			startDate time.Time,
		) (notification.Notifications, error) {
			return notification.Notifications{}, nil
		},
		GetterTypes: func() (notification.Types, error) {
			return []notification.Type{{ID: 1, Name: "type_test", RateLimit: 60, RequestLimit: 1, IsEnable: true}}, nil
		},
		Setter: func(notifType notification.Type, userID int) error { return nil },
	}
}
