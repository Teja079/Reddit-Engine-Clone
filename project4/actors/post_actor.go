package actors

import (
	"fmt"
	"time"

	"github.com/asynkron/protoactor-go/actor"
	"github.com/tejasriramparvathaneni/reddit_clone/proto"
)

type PostActor struct {
	PostID        string
	Content       string
	Author        string
	SubredditName string
	Timestamp     int64
	Comments      map[string]*actor.PID
	Upvotes       int32
	Downvotes     int32
	EnginePID     *actor.PID
}

func NewPostActor(postID, content, author, subredditName string, enginePID *actor.PID) actor.Actor {
	return &PostActor{
		PostID:        postID,
		Content:       content,
		Author:        author,
		SubredditName: subredditName,
		Timestamp:     time.Now().Unix(),
		Comments:      make(map[string]*actor.PID),
		Upvotes:       0,
		Downvotes:     0,
		EnginePID:     enginePID,
	}
}

func (state *PostActor) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *proto.CommentOnPost:
		state.handleCommentOnPost(context, msg)
	case *proto.VoteOnPost:
		state.handleVoteOnPost(context, msg)
	case *proto.CommentOnComment:
		state.forwardCommentOnComment(context, msg)
	case *proto.GetPostDetails:
		state.handleGetPostDetails(context)
	default:
		fmt.Printf("PostActor received a message: %T\n", msg)
	}
}

func (state *PostActor) handleGetPostDetails(context actor.Context) {
	post := &proto.Post{
		Content:       state.Content,
		Author:        state.Author,
		SubredditName: state.SubredditName,
		Timestamp:     state.Timestamp,
		Upvotes:       state.Upvotes,
		Downvotes:     state.Downvotes,
		PostId:        state.PostID,
	}
	context.Respond(post)
}

func (state *PostActor) handleCommentOnPost(context actor.Context, msg *proto.CommentOnPost) {
	commentID := fmt.Sprintf("%s_%d", state.PostID, len(state.Comments)+1)

	commentProps := actor.PropsFromProducer(func() actor.Actor {
		return NewCommentActor(msg.Content, msg.Author, commentID)
	})
	commentPID := context.Spawn(commentProps)
	state.Comments[commentID] = commentPID

	fmt.Printf("Client %s commented on post %s\n", msg.Author, state.PostID)
}

func (state *PostActor) forwardCommentOnComment(context actor.Context, msg *proto.CommentOnComment) {
	parentCommentPID, exists := state.Comments[msg.ParentCommentId]
	if !exists {
		fmt.Printf("Parent comment %s not found for reply by %s\n", msg.ParentCommentId, msg.Author)
		return
	}

	context.Send(parentCommentPID, msg)
}

func (state *PostActor) handleVoteOnPost(context actor.Context, msg *proto.VoteOnPost) {
	if msg.Upvote {
		state.Upvotes++
		state.notifyAuthorKarma(context, 1)
	} else {
		state.Downvotes++
		state.notifyAuthorKarma(context, -1)
	}
	fmt.Printf("Client %s voted on post %s by %s\n", msg.Voter, state.PostID, state.Author)
}

func (state *PostActor) notifyAuthorKarma(context actor.Context, amount int32) {
	updateKarmaMsg := &proto.UpdateKarma{
		Username: state.Author,
		Amount:   amount,
	}
	context.Send(state.EnginePID, updateKarmaMsg)
}
