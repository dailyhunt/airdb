package raft

import (
	"github.com/coreos/etcd/raft/raftpb"
	"github.com/dailyhunt/airdb/proto"
)

func EntriesToString(entries []raftpb.Entry) []string {
	parsed := make([]string, len(entries))

	for i, v := range entries {
		entryType := v.Type
		if entryType == raftpb.EntryNormal && v.Data != nil {
			var p proto.Put
			p.Unmarshal(v.Data)
			parsed[i] = p.String()
		} else {
			parsed[i] = v.String()
		}
	}
	return parsed
}
