package core

import (
	"context"
	"log"
	"math/rand"
	"time"
)

func HeatBeat(ctx context.Context, rf *Raft) error {
	for !rf.killed() {
		log.Printf("pid:%d term:%d role:%d", rf.me, rf.currentTerm, rf.role)
		time.Sleep(time.Duration(150+rand.Intn(200)) * time.Millisecond)
		rf.mu.Lock()

		now := time.Now()
		limit := time.Duration(150+rand.Intn(200)) * time.Millisecond
		elapses := now.Sub(rf.lastActiveTime)
		if rf.role == Follower && elapses > limit {
			log.Println(rf.me, " become candidate")
			rf.role = Candidate
		}
		if rf.role == Candidate {
			rf.election(ctx)
		}
		rf.mu.Unlock()
	}
	log.Println("done")
	return nil
}
