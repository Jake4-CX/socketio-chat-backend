package initializers

import (
	"slices"
	"time"

	"github.com/Jake4-CX/socketio-chat-backend/pkg/structs"
	socketio "github.com/doquangtan/socket.io/v4"
	log "github.com/sirupsen/logrus"
)

var SocketIO *socketio.Io

var ConnectedUsers []*structs.User

func InitializeWebsocket() {
	SocketIO = socketio.New()

	SocketIO.OnConnection(onConnection)
}

func onConnection(socket *socketio.Socket) {
	log.Info("New connection from: ", socket.Id)
	socket.Emit("message", "Hello, world!")

	// We need the user's username. The usernames must be unique (no other user connected can have the same username)
	socket.On("connect", func(event *socketio.EventPayload) {
		if len(event.Data) > 0 && event.Data[0] != nil {
			username := event.Data[0].(string)

			// Check if the username is already taken
			for _, user := range ConnectedUsers {
				if user.Username == username {
					socket.Emit("error", "Username is already taken")
					return
				}
			}

			// Add the user to the list of connected users
			user := &structs.User{
				SocketId:    socket.Id,
				Username:    username,
				ConnectedAt: time.Now(),
			}

			ConnectedUsers = append(ConnectedUsers, user)
		}
	})

	socket.On("join-room", func(event *socketio.EventPayload) {

		if len(socket.Rooms()) > 0 {
			socket.Emit("error", "You cannot be in more than 1 rooms at a time")
			return
		}

		if len(event.Data) > 0 && event.Data[0] != nil {
			room := event.Data[0].(string)
			log.Info("Join room: ", room)
			socket.Join(room)
		}
	})

	socket.On("leave-room", func(event *socketio.EventPayload) {
		if len(event.Data) > 0 && event.Data[0] != nil {
			room := event.Data[0].(string)

			// Check if the user is in the room
			rooms := socket.Rooms()
			if slices.Contains(rooms, room) {
				socket.Emit("error", "You are not in the room")
				return
			}

			log.Info("Leave room: ", room)
			socket.Leave(room)
		}
	})

	socket.On("message", func(event *socketio.EventPayload) {
		if len(event.Data) > 0 && event.Data[0] != nil {
			message := event.Data[0].(string)
			log.Info("Message: ", message)

			// get the room the user is in
			rooms := socket.Rooms()
			for _, room := range rooms {
				// broadcast the message to all users in the room
				socket.To(room).Emit("message", message)
			}

			socket.Emit("message", message)
		}
	})

	socket.On("disconnect", func(event *socketio.EventPayload) {
		log.Info("Disconnect: ", socket.Id)

		// Remove the user from the list of connected users
		for i, user := range ConnectedUsers {
			if user.SocketId == socket.Id {
				ConnectedUsers = append(ConnectedUsers[:i], ConnectedUsers[i+1:]...)
				break
			}
		}
	})
}
