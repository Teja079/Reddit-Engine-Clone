package models

import "github.com/asynkron/protoactor-go/actor"

type Comment struct {
	Content string
	Author  string
	PID     *actor.PID
}
