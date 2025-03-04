package initializers

import (
	"errors"
	"slices"
	"time"

	"github.com/Jake4-CX/socketio-chat-backend/pkg/structs"
	socketio "github.com/doquangtan/socket.io/v4"
	log "github.com/sirupsen/logrus"
)

var SocketIO *socketio.Io

var ConnectedUsers []*structs.User
var AvailableRooms []string = []string{"room1", "room2", "room3"}

func InitializeWebsocket() {
	SocketIO = socketio.New()

	SocketIO.OnConnection(onConnection)
}

func getConnectedUser(socketId string) (*structs.User, error) {
	for _, user := range ConnectedUsers {
		if user.SocketId == socketId {
			return user, nil
		}
	}

	return nil, errors.New("user not found")
}

func onConnection(socket *socketio.Socket) {
	log.Info("New connection from: ", socket.Id)

	socket.On("authenticate", func(event *socketio.EventPayload) {
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

			authenticatedResponse := &structs.AuthenticatedResponse{
				User:  user,
				Rooms: AvailableRooms,
			}

			socket.Emit("authenticated", authenticatedResponse)
		}
	})

	socket.On("joinRoom", func(event *socketio.EventPayload) {

		if len(socket.Rooms()) > 0 {
			socket.Emit("error", "You cannot be in more than 1 rooms at a time")
			return
		}

		if len(event.Data) > 0 && event.Data[0] != nil {
			room := event.Data[0].(string)

			// check if the room exists
			if !slices.Contains(AvailableRooms, room) {
				socket.Emit("error", "Room does not exist")
				return
			}

			socket.Join(room)
			socket.Emit("joinedRoom", room)
			log.Info("User joined room: ", room)
		}
	})

	socket.On("leaveRoom", func(event *socketio.EventPayload) {

		if len(socket.Rooms()) == 0 {
			socket.Emit("error", "You are not in any rooms")
			return
		}

		// Make sure the user is in the room
		foundRoom := ""

		for _, room := range socket.Rooms() {
			if slices.Contains(AvailableRooms, room) {
				foundRoom = room
				break
			}
		}

		if foundRoom == "" {
			socket.Emit("error", "You are not in the room")
			return
		}

		socket.Leave(foundRoom)
		socket.Emit("leftRoom", foundRoom)
	})

	socket.On("sendMessage", func(event *socketio.EventPayload) {

		// Check if the user is authenticated
		user, err := getConnectedUser(socket.Id)
		if err != nil {
			socket.Emit("error", "User is not authenticated")
			return
		}

		if len(event.Data) > 0 && event.Data[0] != nil {
			message := event.Data[0].(string)

			messageResponse := &structs.MessageResponse{
				User:        user,
				Message:     message,
				ProcessedAt: time.Now(),
			}

			// send message to all rooms that the user is in
			rooms := socket.Rooms()
			for _, room := range rooms {
				socket.To(room).Emit("message", messageResponse)
			}
		}
	})

	socket.On("disconnect", func(event *socketio.EventPayload) {
		log.Info("Disconnect: ", socket.Id)

		// Remove the user from list of connected users
		for i, user := range ConnectedUsers {
			if user.SocketId == socket.Id {
				ConnectedUsers = append(ConnectedUsers[:i], ConnectedUsers[i+1:]...)
				break
			}
		}
	})
}
