package main

import (
	"fmt"

	_ "net/http/pprof"

	"github.com/asynkron/protoactor-go/actor"
	"github.com/asynkron/protoactor-go/remote"
	"github.com/tejasriramparvathaneni/reddit_clone/actors"
)

func main() {
	system := actor.NewActorSystem()
	remoteConfig := remote.Configure("127.0.0.1", 8080)
	remoting := remote.NewRemote(system, remoteConfig)
	remoting.Start()

	engineProps := actor.PropsFromProducer(func() actor.Actor {
		return actors.NewEngineActor()
	})
	enginePID, err := system.Root.SpawnNamed(engineProps, "engine")
	if err != nil {
		fmt.Printf("Failed to spawn engine actor: %v\n", err)
		return
	}

	fmt.Println("Engine is running...")
	fmt.Printf("Engine PID: %v\n", enginePID)

	select {}
}
