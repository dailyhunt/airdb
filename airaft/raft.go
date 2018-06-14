package airaft

type RaftNode struct {
	ID    int      // id for raft session
	join  bool     // flag , if node is joining existing cluster
	peers []string // url of raft peers

}
