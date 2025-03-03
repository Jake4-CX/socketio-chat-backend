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

	// We need the user's username. The usernames must be unique (no other user connected can have the same username)
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

			// return success - the user is now authenticated - user information returned
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
			log.Info("Join room: ", room)

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
		// no body needed

		if len(socket.Rooms()) == 0 {
			socket.Emit("error", "You are not in any rooms")
			return
		}

		// Make sure the user is in the room
		rooms := socket.Rooms()
		foundRoom := ""

		for _, room := range rooms {
			if slices.Contains(AvailableRooms, room) {
				foundRoom = room
				break
			}
		}

		if foundRoom == "" {
			socket.Emit("error", "You are not in the room")
			return
		}


		log.Info("Leave room: ", foundRoom)
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
			log.Info("Message: ", message)

			messageResponse := &structs.MessageResponse{
				User:        user,
				Message:     message,
				ProcessedAt: time.Now(),
			}

			// get the room the user is in
			rooms := socket.Rooms()
			for _, room := range rooms {
				// broadcast the message to all users in the room
				socket.To(room).Emit("message", messageResponse)
			}
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
