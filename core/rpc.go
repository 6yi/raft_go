package core

import (
	"context"
	"log"
	"net"
	"net/http"
	"net/rpc"
)

type Client struct {
	pid  int
	addr string
}

func (clientEnd *Client) Call(ctx context.Context, serviceMethod string, args interface{}, reply interface{}) bool {
	var client *rpc.Client
	var err error

	if client, err = rpc.DialHTTP("tcp", clientEnd.addr); err != nil {
		//log.Println(err)
		return false
	}
	defer client.Close()

	if err = client.Call(serviceMethod, args, reply); err != nil {
		//log.Println(err)
		return false
	}
	return true
}

func initRpcServer(ctx context.Context, rf *Raft) {
	server := rpc.NewServer()
	server.Register(rf)

	var err error
	var listener net.Listener
	if listener, err = net.Listen("tcp", rf.peers[rf.me].addr); err != nil {
		log.Fatal(err)
	}
	if err = http.Serve(listener, server); err != nil {
		log.Fatal(err)
	}
	return
}

func SendVoteRequest(ctx context.Context, clientEnd *Client, args *RequestVoteArgs, reply *RequestVoteReply) bool {
	ok := clientEnd.Call(ctx, "Raft.VoteRpc", args, reply)
	return ok
}

type RequestVoteArgs struct {
	Term        int
	CandidateId int
	//LastLogIndex int
	//LastLogTerm  int
}
type RequestVoteReply struct {
	Term        int
	VoteGranted bool
}
type VoteResult struct {
	peerId int
	resp   *RequestVoteReply
}
