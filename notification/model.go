package notification

import (
	"errors"
	"strings"
	"time"
)

type Request struct {
	Type    string
	Message string
	UserID  int
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

type Notifications []Notification
type Notification struct {
	ID     int
	UserID int
	Type   Type
	Date   time.Time
}

func NewNotification(userID int, notifType Type) Notification {
	return Notification{
		UserID: userID,
		Type:   notifType,
	}
}

func (n Notifications) HasAny() bool {
	return len(n) != 0
}

func (n Notification) Exists() bool {
	return n.ID != 0
}

type Types []Type
type Type struct {
	ID           int
	Name         string
	RateLimit    int
	RequestLimit int
	IsEnable     bool
}

func (t Types) HasAny() bool {
	return len(t) != 0
}

func (t Types) GetByName(nType string) Type {
	for _, t := range t {
		if strings.EqualFold(t.Name, nType) {
			return t
		}
	}
	return Type{}
}

func (t Type) Exists() bool {
	return t.ID != 0
}
