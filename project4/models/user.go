package models

import "github.com/asynkron/protoactor-go/actor"

type User struct {
	Username string
	Password string
	PID      *actor.PID
}
