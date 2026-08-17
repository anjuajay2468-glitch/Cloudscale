package raft

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func SendRequestVote(
	peer string,
	args RequestVoteArgs,
) (RequestVoteReply, error) {

	data, err := json.Marshal(args)
	if err != nil {
		return RequestVoteReply{}, err
	}

	url := fmt.Sprintf("%s/raft/request-vote", peer)

	resp, err := http.Post(
		url,
		"application/json",
		bytes.NewReader(data),
	)

	if err != nil {
		return RequestVoteReply{}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return RequestVoteReply{}, fmt.Errorf(
			"peer returned status %d",
			resp.StatusCode,
		)
	}

	var reply RequestVoteReply

	if err := json.NewDecoder(resp.Body).Decode(&reply); err != nil {
		return RequestVoteReply{}, err
	}

	return reply, nil
}