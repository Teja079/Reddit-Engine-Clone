package models

import "github.com/asynkron/protoactor-go/actor"

type Subreddit struct {
	Name string
	PID  *actor.PID
}
