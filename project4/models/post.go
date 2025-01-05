package models

import "github.com/asynkron/protoactor-go/actor"

type Post struct {
	Content string
	Author  string
	PID     *actor.PID
}
