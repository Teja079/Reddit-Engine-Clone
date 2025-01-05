package actors

import (
	"fmt"
	"time"

	"github.com/asynkron/protoactor-go/actor"
	"github.com/tejasriramparvathaneni/reddit_clone/proto"
)

type UserActor struct {
	Username      string
	Karma         int32
	Inbox         []*proto.DirectMessage
	Subscriptions map[string]*actor.PID
}

func NewUserActor(username string) actor.Actor {
	return &UserActor{
		Username:      username,
		Karma:         0,
		Inbox:         []*proto.DirectMessage{},
		Subscriptions: make(map[string]*actor.PID),
	}
}

func (state *UserActor) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *proto.NewPostNotification:
		fmt.Printf("Client %s received new post notification from subreddit %s\n", state.Username, msg.SubredditName)
	case *proto.SendDirectMessage:
		state.handleSendDirectMessage(context, msg)
	case *proto.GetInbox:
		state.handleGetInbox(context, msg)
	case *proto.UpdateKarma:
		state.handleUpdateKarma(msg)
	case *proto.GetFeed:
		state.handleGetFeed(context, msg)
	case *proto.JoinSubreddit:
		state.handleJoinSubreddit(msg)
	case *proto.LeaveSubreddit:
		state.handleLeaveSubreddit(msg)
	default:
		fmt.Printf("UserActor received a message: %T\n", msg)
	}
}

func (state *UserActor) handleUpdateKarma(msg *proto.UpdateKarma) {
	state.Karma += msg.Amount
	fmt.Printf("Client %s karma updated to %d\n", state.Username, state.Karma)
}

func (state *UserActor) handleSendDirectMessage(_ actor.Context, msg *proto.SendDirectMessage) {
	directMessage := &proto.DirectMessage{
		FromUsername: msg.FromUsername,
		Content:      msg.Content,
		Timestamp:    time.Now().Unix(),
	}
	state.Inbox = append(state.Inbox, directMessage)
	fmt.Printf("Client %s received a direct message from %s\n", state.Username, msg.FromUsername)
}

func (state *UserActor) handleGetInbox(context actor.Context, _ *proto.GetInbox) {
	inbox := &proto.Inbox{
		Messages: state.Inbox,
	}
	context.Respond(inbox)
}

func (state *UserActor) handleGetFeed(context actor.Context, _ *proto.GetFeed) {
	var posts []*proto.Post

	for _, subredditPID := range state.Subscriptions {
		if subredditPID != nil {
			future := context.RequestFuture(subredditPID, &proto.GetSubredditPosts{}, 5*time.Second)
			result, err := future.Result()
			if err == nil {
				subredditPosts := result.(*proto.SubredditPosts).Posts
				posts = append(posts, subredditPosts...)
			} else {
				// If any subreddit times out, we just skip it. Could log error.
			}
		}
	}

	feed := &proto.Feed{
		Posts: posts,
	}
	context.Respond(feed)
}

func (state *UserActor) handleJoinSubreddit(msg *proto.JoinSubreddit) {
	subredditPID := actor.NewPID(msg.SubredditPid.Address, msg.SubredditPid.Id)
	state.Subscriptions[msg.SubredditName] = subredditPID
	fmt.Printf("Client %s subscribed to subreddit %s\n", state.Username, msg.SubredditName)
}

func (state *UserActor) handleLeaveSubreddit(msg *proto.LeaveSubreddit) {
	delete(state.Subscriptions, msg.SubredditName)
}
