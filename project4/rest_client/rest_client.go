package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func RunClient() {
	// Register a user
	registerReq := map[string]string{
		"username": "TEJA",
		"password": "password123",
	}
	doPost("http://localhost:3000/users", registerReq)

	// Create a subreddit
	subredditReq := map[string]string{
		"name": "golang",
	}
	doPost("http://localhost:3000/subreddits", subredditReq)

	// Join the subreddit
	joinReq := map[string]string{"username": "TEJA"}
	doPost("http://localhost:3000/subreddits/golang/join", joinReq)

	// Post to the subreddit
	postReq := map[string]string{
		"author":  "TEJA",
		"content": "Hello Gators!",
	}
	doPost("http://localhost:3000/subreddits/golang/posts", postReq)

	// Get feed
	doGet("http://localhost:3000/users/TEJA/feed")

	// Send a direct message
	msgReq := map[string]string{
		"from_username": "TEJA",
		"to_username":   "SATWIK",
		"content":       "Hey SATHWIK!",
	}
	doPost("http://localhost:3000/messages", msgReq)

	// Get inbox
	doGet("http://localhost:3000/users/TEJA/inbox")
}

func doPost(url string, data interface{}) {
	jsonData, _ := json.Marshal(data)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("POST %s error: %v\n", url, err)
		return
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Printf("POST %s -> %s\n", url, string(body))
}

func doGet(url string) {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("GET %s error: %v\n", url, err)
		return
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Printf("GET %s -> %s\n", url, string(body))
}
