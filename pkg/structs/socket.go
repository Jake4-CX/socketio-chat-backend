package structs

import (
	"time"
)

type User struct {
	SocketId    string       `json:"socket_id"`
	Username    string    `json:"username"`
	ConnectedAt time.Time `json:"connected_at"`
}
