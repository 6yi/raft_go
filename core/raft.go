package core

import (
	"context"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

type LogEntry struct {
	Command interface{}
	Term    int
}
type Role int

const (
	Leader Role = iota + 1
	Follower
	Candidate
)

type Raft struct {
	mu    sync.Mutex
	peers []*Client //记录ip，然后调用内部的CALL方法调用RAFT的函数
	me    int
	dead  int32

	currentTerm int // 见过的最大任期
	votedFor    int // 记录在currentTerm任期投票给谁了

	log               []LogEntry
	lastIncludedIndex int
	lastIncludedTerm  int

	commitIndex int // 已知的最大已提交索引
	lastApplied int // 当前应用到状态机的索引

	role              Role      // 1:leader 2:follower 3:candidates
	leaderId          int       // leader的id
	lastActiveTime    time.Time // 上次活跃时间（刷新时机：收到leader心跳、给其他candidates投票、请求其他节点投票）
	lastBroadcastTime time.Time // 作为leader，上次的广播时间

}

func NewRaft(addrs map[int]string, me int) *Raft {
	rf := &Raft{}
	rf.role = Follower
	rf.initConn(addrs)
	rf.me = me
	rf.votedFor = -1
	rf.currentTerm = 0
	rf.lastIncludedIndex = 0
	rf.lastIncludedTerm = -1
	rf.votedFor = -1
	rf.lastActiveTime = time.Now()
	go initRpcServer(context.Background(), rf)
	return rf
}

func (rf *Raft) initConn(addrs map[int]string) {
	peers := make([]*Client, 0)
	for key, addr := range addrs {
		peers = append(peers, &Client{addr: addr, pid: key})
	}
	rf.peers = peers
	return
}
func (rf *Raft) killed() bool {
	z := atomic.LoadInt32(&rf.dead)
	return z == 1
}

func (rf *Raft) election(ctx context.Context) {
	now := time.Now()
	rf.role = Candidate
	rf.lastActiveTime = now
	rf.currentTerm = rf.currentTerm + 1
	rf.votedFor = rf.me
	args := &RequestVoteArgs{
		Term:        rf.currentTerm,
		CandidateId: rf.me,
	}
	voteChan := make(chan *VoteResult, len(rf.peers))
	finishVote := 0
	for _, client := range rf.peers {
		go func(client *Client) {
			if client.pid == rf.me {
				return
			}
			reply := &RequestVoteReply{}
			if SendVoteRequest(ctx, client, args, reply) {
				voteChan <- &VoteResult{peerId: client.pid, resp: reply}
			} else {
				voteChan <- &VoteResult{}
			}
			finishVote++
		}(client)
	}

	voteCount := 1 // 默认自己一票
	maxTerm := 0
	for {
		select {
		case reply := <-voteChan:
			if reply.resp != nil {
				if reply.resp.Term > maxTerm {
					maxTerm = reply.resp.Term
				}
				if reply.resp.VoteGranted {
					voteCount++
				}
			}
		case <-time.After(time.Second * 1):
			goto VOTE_END
		}
		if finishVote == len(rf.peers) || voteCount > len(rf.peers)/2 {
			goto VOTE_END
		}
	}

VOTE_END:
	log.Println(rf.me, " finish vote,count ", voteCount)
	if maxTerm > rf.currentTerm {
		rf.role = Follower
		rf.currentTerm = maxTerm
		rf.votedFor = -1
	}
	if voteCount > len(rf.peers)/2 {
		log.Println(rf.me, " become leader in ", rf.currentTerm, " term")
		rf.role = Leader
		rf.leaderId = rf.me
		rf.lastBroadcastTime = time.Unix(0, 0) // 令appendEntries广播立即执行
		return
	}
}

func (rf *Raft) VoteRpc(args *RequestVoteArgs, reply *RequestVoteReply) error {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	reply.VoteGranted = false
	reply.Term = rf.currentTerm
	if args.Term < rf.currentTerm {
		return nil
	}
	if args.Term > rf.currentTerm {
		rf.currentTerm = args.Term
		rf.votedFor = -1
		rf.role = Follower
		rf.leaderId = -1
	}
	if rf.votedFor == -1 {
		rf.votedFor = args.CandidateId
		rf.lastActiveTime = time.Now()

		reply.VoteGranted = true
	}
	return nil
}
