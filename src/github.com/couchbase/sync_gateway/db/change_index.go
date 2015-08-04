package db

import (
	"github.com/couchbase/sync_gateway/base"
	"github.com/couchbase/sync_gateway/channels"
)

// A ChangeIndex is responsible for indexing incoming events from change_listener, and
// and servicing requests for indexed changes from changes.go.  In addition to basic feed processiw2BV23                                                                                                                                                                                                                                                                                                                                                                                                                                                                             ng, the
// ChangeIndex is the component responsible for index consistency and sequence management.  The ChangeIndex
// doesn't define storage - see ChannelIndex for storage details.
//
// Currently there are two ChangeIndex implementations:
//   1. change_cache.go.
//      Stores recent index entries in memory, retrieves older entries
//      using view query.
//      Assumes a single SG node (itself) as index writer, processing entire mutation stream.
//      Uses _sync:seq for sequence management, and buffers
//      incoming sequences from the feed to provide consistency.
//   2. kv_change_index.go
//      Supports multiple SG nodes as index writers.  Uses vbucket sequence numbers for sequence management,
//      and maintains stable sequence vector clocks to provide consistency.

type ChangeIndex interface {

	// Initialize the index
	Init(context *DatabaseContext, lastSequence SequenceID, onChange func(base.Set), cacheOptions *CacheOptions, indexOptions *ChangeIndexOptions)

	// Stop the index
	Stop()

	// Clear the index
	Clear()

	// Enable/Disable indexing
	EnableChannelIndexing(enable bool)

	// Retrieve changes in a channel
	GetChanges(channelName string, options ChangesOptions) ([]*LogEntry, error)
	// Retrieve in-memory changes in a channel
	GetCachedChanges(channelName string, options ChangesOptions) (validFrom uint64, entries []*LogEntry)

	// Called to add a document to the index
	DocChanged(docID string, docJSON []byte, vbucket uint64)

	// Retrieves stable sequence for index
	GetStableSequence(docID string) SequenceID

	// Utility functions for unit testing
	waitForSequenceID(sequence SequenceID)

	// Handling specific to change_cache.go's sequence handling.  Ideally should refactor usage in changes.go to push
	// down into internal change_cache.go handling, but it's non-trivial refactoring
	getOldestSkippedSequence() uint64
	getChannelCache(channelName string) *channelCache

	// Unit test support
	waitForSequence(sequence uint64)
	waitForSequenceWithMissing(sequence uint64)
}

// Index type
type IndexType uint8

const (
	KVIndex IndexType = iota
	MemoryCache
)

type ChangeIndexOptions struct {
	Type    IndexType    // Index type
	Bucket  base.Bucket  // Indexing bucket
	Options CacheOptions // Caching options
}

// ChannelIndex defines the API used by the ChangeIndex to interact with the underlying index storage
type ChannelIndex interface {
	Add(entry *LogEntry) error
	AddSet(entries []*LogEntry) error
	GetClock() (uint64, error)
	SetClock() (uint64, error)
	GetCachedChanges(options ChangesOptions, stableSequence uint64)
	Compact()
}

func (entry *LogEntry) isRemoved() bool {
	return entry.Flags&channels.Removed != 0
}

func (entry *LogEntry) setRemoved() {
	entry.Flags |= channels.Removed
}
