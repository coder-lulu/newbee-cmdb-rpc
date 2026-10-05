package cache

import (
	"testing"
	"time"
)

func TestGetStatsReturnsIndependentSnapshot(t *testing.T) {
	reset := time.Now()
	manager := &SmartCacheManager{stats: &CacheStats{
		LocalHits: 1, LocalMisses: 2, DistributedHits: 3,
		DistributedMisses: 4, TotalRequests: 10, LastResetTime: reset,
	}}
	snapshot := manager.GetStats()
	if snapshot.LocalHits != 1 || snapshot.LocalMisses != 2 || snapshot.DistributedHits != 3 ||
		snapshot.DistributedMisses != 4 || snapshot.TotalRequests != 10 || !snapshot.LastResetTime.Equal(reset) {
		t.Fatal("snapshot lost statistics")
	}
	if !snapshot.mu.TryLock() {
		t.Fatal("snapshot retained the source read lock")
	}
	snapshot.mu.Unlock()
	manager.ResetStats()
	if snapshot.TotalRequests != 10 || snapshot.LocalHits != 1 {
		t.Fatal("reset mutated an existing snapshot")
	}
	snapshot.LocalHits = 99
	if manager.GetStats().LocalHits != 0 {
		t.Fatal("snapshot mutation changed the source")
	}
}
