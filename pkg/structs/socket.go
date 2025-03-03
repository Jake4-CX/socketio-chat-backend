package structs

import (
	"time"
)

type User struct {
	SocketId    string    `json:"socket_id"`
	Username    string    `json:"username"`
	ConnectedAt time.Time `json:"connected_at"`
}

type AuthenticatedResponse struct {
	User  *User    `json:"user"`
	Rooms []string `json:"rooms"`
}

type MessageResponse struct {
	User        *User     `json:"user"`
	Message     string    `json:"message"`
	ProcessedAt time.Time `json:"processed_at"`
}
