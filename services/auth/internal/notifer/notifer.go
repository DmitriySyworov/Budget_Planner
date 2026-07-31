package notifer

import (
	"app/auth-service/internal/common"
	"context"
	"shared/loggers"
	"shared/shconstant"
	"shared/shkafka"
	"shared/shprotos/event"
	"time"

	"google.golang.org/protobuf/proto"
)

type Notifer struct {
	Logger            *loggers.Logger
	ProduceEmailEvent *shkafka.KafkaProducer
}

func NewNotifer(producer *shkafka.KafkaProducer, logger *loggers.Logger) *Notifer {
	return &Notifer{
		Logger:            logger,
		ProduceEmailEvent: producer,
	}
}

type NotificationEvent struct {
	*LetterAuth
	*LetterSecurity
	EmailTo   string
	EventUUID string
}

type LetterAuth struct {
	ValidTime int64
	Code      string
}

type LetterSecurity struct {
	Device      string
	IP          string
	CurrentTime string
}
type NotificationAction string

const (
	SendCodeAction          NotificationAction = "code_action"
	SendNewDeviceAction     NotificationAction = "new_device_action"
	SendSecurityAlertAction NotificationAction = "security_alert_action"
)

func (n *Notifer) HelperSendEmailEvent(ntEvent *NotificationEvent, action NotificationAction) {
	var dataEvent []byte
	var errMarshal error
	switch action {
	case SendCodeAction:
		eventSendEmail := &event.NotificationEvent{
			Event: &event.NotificationEvent_AuthCode{
				AuthCode: &event.AuthLetterPayload{
					ValidTime: time.Now().Add(common.TTLSessionJWT).Unix(),
					Code:      ntEvent.Code,
				},
			},
			EmailTo:   ntEvent.EmailTo,
			EventUUID: ntEvent.EventUUID,
		}
		dataEvent, errMarshal = proto.Marshal(eventSendEmail)
		if errMarshal != nil {
			n.Logger.Error("failed to marshal proto event email send: " + errMarshal.Error())
			return
		}
	case SendNewDeviceAction:
		eventNewDevice := &event.NotificationEvent{
			Event: &event.NotificationEvent_NewDevice{
				NewDevice: &event.SecurityLetterPayload{
					Device:      ntEvent.LetterSecurity.Device,
					ClientIp:    ntEvent.LetterSecurity.IP,
					CurrentTime: time.Now().Format(time.DateTime),
				},
			},
			EmailTo:   ntEvent.EmailTo,
			EventUUID: ntEvent.EventUUID,
		}
		dataEvent, errMarshal = proto.Marshal(eventNewDevice)
		if errMarshal != nil {
			n.Logger.Error("failed to marshal proto event security alert letter: " + errMarshal.Error())
			return
		}
	case SendSecurityAlertAction:
		eventSecurityAlert := &event.NotificationEvent{
			Event: &event.NotificationEvent_SecurityAlert{
				SecurityAlert: &event.SecurityLetterPayload{
					Device:      ntEvent.LetterSecurity.Device,
					ClientIp:    ntEvent.LetterSecurity.IP,
					CurrentTime: time.Now().Format(time.DateTime),
				},
			},
			EmailTo:   ntEvent.EmailTo,
			EventUUID: ntEvent.EventUUID,
		}
		dataEvent, errMarshal = proto.Marshal(eventSecurityAlert)
		if errMarshal != nil {
			n.Logger.Error("failed to marshal proto event security alert letter: " + errMarshal.Error())
			return
		}
	}
	ctxTimeout, cancel := context.WithTimeout(context.Background(), shconstant.CtxTimeoutSendEventKafka)
	defer cancel()
	if errSendEvent := n.ProduceEmailEvent.SendEvent(ctxTimeout, ntEvent.EventUUID, dataEvent); errSendEvent != nil {
		n.Logger.Error("failed to send event email letter: " + errSendEvent.Error())
	}
}
