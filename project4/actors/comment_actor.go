package actors

import (
	"fmt"
	"time"

	"github.com/asynkron/protoactor-go/actor"
	"github.com/tejasriramparvathaneni/reddit_clone/proto"
)

type CommentActor struct {
	Content   string
	Author    string
	Timestamp int64
	Upvotes   int32
	Downvotes int32
	Replies   map[string]*actor.PID
	CommentID string
}

func NewCommentActor(content, author, commentID string) actor.Actor {
	return &CommentActor{
		Content:   content,
		Author:    author,
		Timestamp: time.Now().Unix(),
		Upvotes:   0,
		Downvotes: 0,
		Replies:   make(map[string]*actor.PID),
		CommentID: commentID,
	}
}

func (state *CommentActor) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *proto.CommentOnComment:
		state.handleCommentOnComment(context, msg)
	case *proto.VoteOnComment:
		state.handleVoteOnComment(context, msg)
	default:
		fmt.Printf("CommentActor received a message: %T\n", msg)
	}
}

func (state *CommentActor) handleCommentOnComment(context actor.Context, msg *proto.CommentOnComment) {
	replyCommentID := fmt.Sprintf("%s_%d", state.CommentID, len(state.Replies)+1)

	replyProps := actor.PropsFromProducer(func() actor.Actor {
		return NewCommentActor(msg.Content, msg.Author, replyCommentID)
	})
	replyPID := context.Spawn(replyProps)
	state.Replies[replyCommentID] = replyPID

	fmt.Printf("Client %s replied to comment %s by %s\n", msg.Author, state.CommentID, state.Author)
}

func (state *CommentActor) handleVoteOnComment(_ actor.Context, msg *proto.VoteOnComment) {
	if msg.Upvote {
		state.Upvotes++
	} else {
		state.Downvotes++
	}
	fmt.Printf("Client %s voted on comment %s by %s\n", msg.Voter, state.CommentID, state.Author)
}
