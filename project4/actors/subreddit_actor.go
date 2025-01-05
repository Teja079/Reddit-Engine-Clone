package actors

import (
	"fmt"
	"time"

	"github.com/asynkron/protoactor-go/actor"
	"github.com/tejasriramparvathaneni/reddit_clone/proto"
)

type SubredditActor struct {
	Name      string
	Members   map[string]*actor.PID
	Posts     map[string]*actor.PID
	EnginePID *actor.PID
}

func NewSubredditActor(name string, enginePID *actor.PID) actor.Actor {
	return &SubredditActor{
		Name:      name,
		Members:   make(map[string]*actor.PID),
		Posts:     make(map[string]*actor.PID),
		EnginePID: enginePID,
	}
}

func (state *SubredditActor) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *proto.JoinSubreddit:
		state.handleJoinSubreddit(context, msg)
	case *proto.LeaveSubreddit:
		state.handleLeaveSubreddit(context, msg)
	case *proto.PostToSubreddit:
		state.handlePostToSubreddit(context, msg)
	case *proto.GetSubredditPosts:
		state.handleGetSubredditPosts(context)
	default:
		fmt.Printf("SubredditActor received unknown message: %T\n", msg)
	}
}

func (state *SubredditActor) handleJoinSubreddit(context actor.Context, msg *proto.JoinSubreddit) {
	userPID := actor.NewPID(msg.UserPid.Address, msg.UserPid.Id)
	state.Members[msg.Username] = userPID
	fmt.Printf("Client %s joined subreddit %s\n", msg.Username, state.Name)

	// Notify UserActor about the subscription
	context.Send(userPID, msg)
}

func (state *SubredditActor) handleLeaveSubreddit(_ actor.Context, msg *proto.LeaveSubreddit) {
	delete(state.Members, msg.Username)
	fmt.Printf("Client %s left subreddit %s\n", msg.Username, state.Name)
}

func (state *SubredditActor) handlePostToSubreddit(context actor.Context, msg *proto.PostToSubreddit) {
	// Generate a unique post ID
	postID := fmt.Sprintf("%s_%d", state.Name, len(state.Posts)+1)

	postProps := actor.PropsFromProducer(func() actor.Actor {
		return NewPostActor(postID, msg.Content, msg.Author, state.Name, state.EnginePID)
	})
	postPID := context.Spawn(postProps)
	state.Posts[postID] = postPID

	fmt.Printf("Client %s posted to subreddit %s\n", msg.Author, state.Name)

	// Notify members
	for _, userPID := range state.Members {
		notification := &proto.NewPostNotification{
			SubredditName: state.Name,
			PostId:        postID,
			Content:       msg.Content,
			Author:        msg.Author,
		}
		context.Send(userPID, notification)
	}
}

func (state *SubredditActor) handleGetSubredditPosts(context actor.Context) {
	var posts []*proto.Post
	for _, postPID := range state.Posts {
		// Request post details from PostActor
		future := context.RequestFuture(postPID, &proto.GetPostDetails{}, 2*time.Second)
		result, err := future.Result()
		if err == nil {
			post := result.(*proto.Post)
			posts = append(posts, post)
		}
	}
	response := &proto.SubredditPosts{
		Posts: posts,
	}
	context.Respond(response)
}
