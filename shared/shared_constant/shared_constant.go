package shared_constant

import "time"

const (
	TypeSoftDelete       = "soft-delete"
	TypeHardDelete       = "hard-delete"
	CtxTimeoutRedis      = time.Second * 5
	EventDeletedUserUUID = "deleted_user_uuid"
)
