package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/asynkron/protoactor-go/actor"
	"github.com/asynkron/protoactor-go/remote"
	"github.com/gin-gonic/gin"
	"github.com/tejasriramparvathaneni/reddit_clone/actors"
	"github.com/tejasriramparvathaneni/reddit_clone/proto"
)

var (
	system    *actor.ActorSystem
	enginePID *actor.PID
)

func StartServer() {
	system = actor.NewActorSystem()
	// Use a different port than engine to avoid conflicts, e.g. 8081
	remoteConfig := remote.Configure("127.0.0.1", 8081)
	remoting := remote.NewRemote(system, remoteConfig)
	remoting.Start()

	engineProps := actor.PropsFromProducer(func() actor.Actor {
		return actors.NewEngineActor()
	})
	var err error
	enginePID, err = system.Root.SpawnNamed(engineProps, "engine")
	if err != nil {
		log.Fatalf("Failed to spawn engine actor: %v\n", err)
	}

	fmt.Println("Engine is running...")
	fmt.Printf("Engine PID: %v\n", enginePID)

	r := gin.Default()

	// Define all required routes
	r.POST("/users", registerUserHandler)
	r.GET("/users/:username/inbox", getInboxHandler)
	r.GET("/users/:username/feed", getFeedHandler)

	r.POST("/subreddits", createSubredditHandler)
	r.POST("/subreddits/:name/join", joinSubredditHandler)
	r.POST("/subreddits/:name/leave", leaveSubredditHandler)
	r.POST("/subreddits/:name/posts", postToSubredditHandler)

	r.POST("/posts/:post_id/comments", commentOnPostHandler)
	r.POST("/posts/:post_id/votes", voteOnPostHandler)

	r.POST("/messages", sendDirectMessageHandler)

	err = r.Run(":3000")
	if err != nil {
		fmt.Println("Failed to start server:", err)
	}
}

type RegisterUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func registerUserHandler(c *gin.Context) {
	var req RegisterUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	future := system.Root.RequestFuture(enginePID, &proto.RegisterUser{
		Username: req.Username,
		Password: req.Password,
	}, 5*time.Second)

	result, err := future.Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Engine timeout or error"})
		return
	}

	resp := result.(*proto.RegistrationResponse)
	if !resp.Success {
		c.JSON(http.StatusConflict, gin.H{"message": resp.Message})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": resp.Message})
}

func createSubredditHandler(c *gin.Context) {
	var req struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subreddit name"})
		return
	}

	system.Root.Send(enginePID, &proto.CreateSubreddit{Name: req.Name})
	c.JSON(http.StatusOK, gin.H{"message": "Subreddit creation requested"})
}

func joinSubredditHandler(c *gin.Context) {
	subredditName := c.Param("name")
	var req struct {
		Username string `json:"username"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid username"})
		return
	}

	system.Root.Send(enginePID, &proto.JoinSubreddit{
		Username:      req.Username,
		SubredditName: subredditName,
	})

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Join request for %s sent", subredditName)})
}

func leaveSubredditHandler(c *gin.Context) {
	subredditName := c.Param("name")
	var req struct {
		Username string `json:"username"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid username"})
		return
	}

	system.Root.Send(enginePID, &proto.LeaveSubreddit{
		Username:      req.Username,
		SubredditName: subredditName,
	})

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Leave request for %s sent", subredditName)})
}

func postToSubredditHandler(c *gin.Context) {
	subredditName := c.Param("name")
	var req struct {
		Content string `json:"content"`
		Author  string `json:"author"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Author == "" || req.Content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post data"})
		return
	}

	system.Root.Send(enginePID, &proto.PostToSubreddit{
		Content:       req.Content,
		Author:        req.Author,
		SubredditName: subredditName,
	})

	c.JSON(http.StatusOK, gin.H{"message": "Post request sent"})
}

func commentOnPostHandler(c *gin.Context) {
	postID := c.Param("post_id")
	var req struct {
		Content string `json:"content"`
		Author  string `json:"author"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Author == "" || req.Content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid comment data"})
		return
	}

	system.Root.Send(enginePID, &proto.CommentOnPost{
		Content: req.Content,
		Author:  req.Author,
		PostId:  postID,
	})

	c.JSON(http.StatusOK, gin.H{"message": "Comment sent"})
}

func voteOnPostHandler(c *gin.Context) {
	postID := c.Param("post_id")
	var req struct {
		Voter  string `json:"voter"`
		Upvote bool   `json:"upvote"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Voter == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid vote data"})
		return
	}

	system.Root.Send(enginePID, &proto.VoteOnPost{
		PostId: postID,
		Upvote: req.Upvote,
		Voter:  req.Voter,
	})

	c.JSON(http.StatusOK, gin.H{"message": "Vote submitted"})
}

func sendDirectMessageHandler(c *gin.Context) {
	var req struct {
		FromUsername string `json:"from_username"`
		ToUsername   string `json:"to_username"`
		Content      string `json:"content"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.FromUsername == "" || req.ToUsername == "" || req.Content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message data"})
		return
	}

	system.Root.Send(enginePID, &proto.SendDirectMessage{
		FromUsername: req.FromUsername,
		ToUsername:   req.ToUsername,
		Content:      req.Content,
	})

	c.JSON(http.StatusOK, gin.H{"message": "Message sent"})
}

func getInboxHandler(c *gin.Context) {
	username := c.Param("username")

	future := system.Root.RequestFuture(enginePID, &proto.GetInbox{
		Username: username,
	}, 5*time.Second)

	result, err := future.Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Timeout or error"})
		return
	}
	inbox := result.(*proto.Inbox)
	c.JSON(http.StatusOK, gin.H{"inbox": inbox.Messages})
}

func getFeedHandler(c *gin.Context) {
	username := c.Param("username")

	future := system.Root.RequestFuture(enginePID, &proto.GetFeed{
		Username: username,
	}, 5*time.Second)

	result, err := future.Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Timeout or error"})
		return
	}
	feed := result.(*proto.Feed)
	c.JSON(http.StatusOK, gin.H{"feed": feed.Posts})
}
