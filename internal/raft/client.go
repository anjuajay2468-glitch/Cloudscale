package raft

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/anjuajay2468-glitch/cloudscale/internal/auth"
)

const internalTokenEnv = "CLOUDSCALE_INTERNAL_TOKEN"

func newInternalClient() *http.Client {
	return &http.Client{
		Timeout: 150 * time.Millisecond,
	}
}

// addInternalAuth adds the node-to-node authentication token when one
// is configured.
//
// The token is optional at the Raft client level because the Raft package
// is also used by unit tests that use unauthenticated httptest servers.
// Production node startup will configure the token.
func addInternalAuth(req *http.Request) {
	token := os.Getenv(internalTokenEnv)

	if token == "" {
		return
	}

	req.Header.Set(
		auth.InternalTokenHeader,
		token,
	)
}

func SendRequestVote(
	peer string,
	args RequestVoteArgs,
) (RequestVoteReply, error) {

	data, err := json.Marshal(args)
	if err != nil {
		return RequestVoteReply{}, err
	}

	url := fmt.Sprintf(
		"%s/raft/request-vote",
		peer,
	)

	req, err := http.NewRequest(
		http.MethodPost,
		url,
		bytes.NewReader(data),
	)
	if err != nil {
		return RequestVoteReply{}, err
	}

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	addInternalAuth(req)

	resp, err := newInternalClient().Do(req)
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

	if err := json.NewDecoder(
		resp.Body,
	).Decode(&reply); err != nil {
		return RequestVoteReply{}, err
	}

	return reply, nil
}

func SendAppendEntries(
	peer string,
	args AppendEntriesArgs,
) (AppendEntriesReply, error) {

	data, err := json.Marshal(args)
	if err != nil {
		return AppendEntriesReply{}, err
	}

	url := fmt.Sprintf(
		"%s/raft/append-entries",
		peer,
	)

	req, err := http.NewRequest(
		http.MethodPost,
		url,
		bytes.NewReader(data),
	)
	if err != nil {
		return AppendEntriesReply{}, err
	}

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	addInternalAuth(req)

	resp, err := newInternalClient().Do(req)
	if err != nil {
		return AppendEntriesReply{}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return AppendEntriesReply{}, fmt.Errorf(
			"peer returned status %d",
			resp.StatusCode,
		)
	}

	var reply AppendEntriesReply

	if err := json.NewDecoder(
		resp.Body,
	).Decode(&reply); err != nil {
		return AppendEntriesReply{}, err
	}

	return reply, nil
}
