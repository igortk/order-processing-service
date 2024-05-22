package handlers

type MessageHandler interface {
	HandleMessage([]byte)
}
