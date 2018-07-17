package raft

import (
	"github.com/coreos/etcd/raft/raftpb"
	"github.com/dailyhunt/airdb/mutation"
)

func EntriesToString(entries []raftpb.Entry) []string {
	parsed := make([]string, len(entries))

	for i, v := range entries {
		entryType := v.Type
		if entryType == raftpb.EntryNormal && v.Data != nil {
			m, err := mutation.Decode(v.Data)
			if err != nil {
				panic(err)
			}
			parsed[i] = m.String()
		} else {
			parsed[i] = v.String()
		}
	}
	return parsed
}
