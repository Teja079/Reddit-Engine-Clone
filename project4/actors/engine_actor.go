package actors

import (
	"fmt"
	"time"

	"github.com/asynkron/protoactor-go/actor"
	log "github.com/sirupsen/logrus"
	"github.com/tejasriramparvathaneni/reddit_clone/models"
	"github.com/tejasriramparvathaneni/reddit_clone/proto"
)

func init() {
	log.SetFormatter(&log.JSONFormatter{})
}

type EngineActor struct {
	users         map[string]*models.User
	subreddits    map[string]*models.Subreddit
	totalMessages int64
}

func NewEngineActor() actor.Actor {
	engine := &EngineActor{
		users:      make(map[string]*models.User),
		subreddits: make(map[string]*models.Subreddit),
	}
	engine.startMetricsLogger()
	return engine
}

func (state *EngineActor) Receive(context actor.Context) {
	state.totalMessages++
	switch msg := context.Message().(type) {
	case *proto.RegisterUser:
		state.handleRegisterUser(context, msg)
	case *proto.CreateSubreddit:
		state.handleCreateSubreddit(context, msg)
	case *proto.JoinSubreddit:
		state.handleJoinSubreddit(context, msg)
	case *proto.LeaveSubreddit:
		state.handleLeaveSubreddit(context, msg)
	case *proto.PostToSubreddit:
		state.handlePostToSubreddit(context, msg)
	case *proto.SendDirectMessage:
		state.handleSendDirectMessage(context, msg)
	case *proto.GetInbox:
		state.handleGetInbox(context, msg)
	case *proto.GetFeed:
		state.handleGetFeed(context, msg)
	case *proto.UpdateKarma:
		state.handleUpdateKarma(context, msg)
	default:
		log.WithFields(log.Fields{
			"actor":   "EngineActor",
			"message": fmt.Sprintf("%T", msg),
		}).Warn("Received a message")
	}
}

func (state *EngineActor) handleRegisterUser(context actor.Context, msg *proto.RegisterUser) {
	if _, exists := state.users[msg.Username]; exists {
		response := &proto.RegistrationResponse{
			Success: false,
			Message: "Username already exists",
		}
		context.Respond(response)
		return
	}

	userProps := actor.PropsFromProducer(func() actor.Actor {
		return NewUserActor(msg.Username)
	})
	userPID := context.Spawn(userProps)

	user := &models.User{
		Username: msg.Username,
		Password: msg.Password,
		PID:      userPID,
	}
	state.users[msg.Username] = user

	response := &proto.RegistrationResponse{
		Success: true,
		Message: "Registration successful",
	}
	context.Respond(response)
}

func (state *EngineActor) handleCreateSubreddit(context actor.Context, msg *proto.CreateSubreddit) {
	if _, exists := state.subreddits[msg.Name]; exists {
		fmt.Printf("Subreddit %s already exists\n", msg.Name)
		return
	}

	subredditProps := actor.PropsFromProducer(func() actor.Actor {
		return NewSubredditActor(msg.Name, context.Self())
	})
	subredditPID := context.Spawn(subredditProps)

	subreddit := &models.Subreddit{
		Name: msg.Name,
		PID:  subredditPID,
	}
	state.subreddits[msg.Name] = subreddit

	fmt.Printf("Subreddit %s created successfully\n", msg.Name)
}

func (state *EngineActor) handleJoinSubreddit(context actor.Context, msg *proto.JoinSubreddit) {
	subreddit, exists := state.subreddits[msg.SubredditName]
	if !exists {
		log.WithFields(log.Fields{
			"subreddit": msg.SubredditName,
		}).Error("Subreddit does not exist")
		return
	}

	user, userExists := state.users[msg.Username]
	if !userExists {
		fmt.Printf("Client %s does not exist\n", msg.Username)
		return
	}

	userPidMessage := &proto.PID{
		Address: user.PID.Address,
		Id:      user.PID.Id,
	}
	msg.UserPid = userPidMessage

	subredditPidMessage := &proto.PID{
		Address: subreddit.PID.Address,
		Id:      subreddit.PID.Id,
	}
	msg.SubredditPid = subredditPidMessage

	context.Send(subreddit.PID, msg)
}

func (state *EngineActor) handleLeaveSubreddit(context actor.Context, msg *proto.LeaveSubreddit) {
	subreddit, exists := state.subreddits[msg.SubredditName]
	if !exists {
		fmt.Printf("Subreddit %s does not exist\n", msg.SubredditName)
		return
	}

	if _, userExists := state.users[msg.Username]; !userExists {
		fmt.Printf("Client %s does not exist\n", msg.Username)
		return
	}

	context.Send(subreddit.PID, msg)
}

func (state *EngineActor) handlePostToSubreddit(context actor.Context, msg *proto.PostToSubreddit) {
	subreddit, exists := state.subreddits[msg.SubredditName]
	if !exists {
		fmt.Printf("Subreddit %s does not exist\n", msg.SubredditName)
		return
	}

	context.Send(subreddit.PID, msg)
}

func (state *EngineActor) handleSendDirectMessage(context actor.Context, msg *proto.SendDirectMessage) {
	recipient, exists := state.users[msg.ToUsername]
	if !exists {
		fmt.Printf("Client %s does not exist\n", msg.ToUsername)
		return
	}

	context.Send(recipient.PID, msg)
}

func (state *EngineActor) handleGetInbox(context actor.Context, msg *proto.GetInbox) {
	user, exists := state.users[msg.Username]
	if !exists {
		fmt.Printf("Client %s does not exist\n", msg.Username)
		inbox := &proto.Inbox{Messages: []*proto.DirectMessage{}}
		context.Respond(inbox)
		return
	}

	future := context.RequestFuture(user.PID, msg, 5*time.Second)
	res, err := future.Result()
	if err != nil {
		fmt.Printf("Error getting inbox for user %s: %v\n", msg.Username, err)
		inbox := &proto.Inbox{Messages: []*proto.DirectMessage{}}
		context.Respond(inbox)
		return
	}
	context.Respond(res)
}

func (state *EngineActor) handleGetFeed(context actor.Context, msg *proto.GetFeed) {
	user, exists := state.users[msg.Username]
	if !exists {
		fmt.Printf("Client %s does not exist\n", msg.Username)
		feed := &proto.Feed{Posts: []*proto.Post{}}
		context.Respond(feed)
		return
	}

	future := context.RequestFuture(user.PID, msg, 5*time.Second)
	res, err := future.Result()
	if err != nil {
		fmt.Printf("Error getting feed for user %s: %v\n", msg.Username, err)
		feed := &proto.Feed{Posts: []*proto.Post{}}
		context.Respond(feed)
		return
	}
	context.Respond(res)
}

func (state *EngineActor) handleUpdateKarma(context actor.Context, msg *proto.UpdateKarma) {
	user, exists := state.users[msg.Username]
	if exists {
		context.Send(user.PID, msg)
	}
}

func (state *EngineActor) startMetricsLogger() {
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		for {
			<-ticker.C
			fmt.Printf("Total messages processed: %d\n", state.totalMessages)
		}
	}()
}
