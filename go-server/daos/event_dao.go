package dao

import (
	"github.com/homingos/campaign-svc/dtos"
)

type EventDao interface {
	ProcessEventDao(eventProp dtos.Event, sourceIp string) error
}
