package orchestrator

import (
	"context"
	"errors"
	"sort"
	"strconv"
	"sync"
	"time"

	"cloudperf/backend/internal/db"
	"cloudperf/backend/internal/model"
	"cloudperf/backend/internal/ws"
)

type Orchestrator struct {
	store *db.Store
	hub   *ws.Hub

	queue chan string

	globalSem chan struct{}

	nodeMu  sync.Mutex
	nodeSem map[string]chan struct{}
}

func New(store *db.Store, hub *ws.Hub) *Orchestrator {
	return &Orchestrator{
		store:     store,
		hub:       hub,
		queue:     make(chan string, 128),
		globalSem: make(chan struct{}, 20),
		nodeSem:   make(map[string]chan struct{}),
	}
}

func (o *Orchestrator) Enqueue(runID string) {
	select {
	case o.queue <- runID:
	default:
	}
}

func (o *Orchestrator) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case runID := <-o.queue:
			go o.executeRun(runID)
		}
	}
}

func (o *Orchestrator) executeRun(runID string) {
	run, err := o.store.GetRun(runID)
	if err != nil || run == nil {
		return
	}
	_ = o.store.UpdateRunStatus(runID, "running", true, false)

	pairs, err := o.store.ListRunPairsForExecution(runID)
	if err != nil {
		_ = o.store.UpdateRunStatus(runID, "failed", false, true)
		return
	}

	var wg sync.WaitGroup
	for _, p := range pairs {
		pair := p
		wg.Add(1)
		go func() {
			defer wg.Done()
			o.runPair(run, pair)
		}()
	}
	wg.Wait()

	updatedPairs, err := o.store.ListRunPairs(runID)
	if err != nil {
		_ = o.store.UpdateRunStatus(runID, "failed", false, true)
		return
	}
	final := aggregateStatus(updatedPairs)
	_ = o.store.UpdateRunStatus(runID, final, false, true)
}

func aggregateStatus(pairs []model.RunPair) string {
	if len(pairs) == 0 {
		return "failed"
	}
	succ, fail := 0, 0
	for _, p := range pairs {
		switch p.Status {
		case "success":
			succ++
		default:
			fail++
		}
	}
	if fail == 0 {
		return "success"
	}
	if succ > 0 {
		return "partial_failed"
	}
	return "failed"
}

func (o *Orchestrator) runPair(run *model.Run, pair model.RunPair) {
	_ = o.store.UpdatePairStatus(pair.ID, "running", "", true, false)
	release := o.acquire(pair.SourceNodeID, pair.TargetNodeID)
	defer release()

	if !o.hub.HasAgent(pair.SourceNodeID) || !o.hub.HasAgent(pair.TargetNodeID) {
		o.failPair(pair.ID, "source or target node is offline")
		return
	}

	port := 5201 + int(time.Now().UnixNano()%300)
	duration := intFromParams(run.Params, "duration", 10)
	parallel := intFromParams(run.Params, "parallel_streams", 1)
	udpBw := strFromParams(run.Params, "udp_bandwidth", "100M")

	targetIP, err := o.store.GetNodeIP(pair.TargetNodeID)
	if err != nil || targetIP == "" {
		o.failPair(pair.ID, "target node public_ip is empty")
		return
	}

	serverReady := false
	for i := 0; i < 3; i++ {
		resp, err := o.hub.SendCommand(pair.TargetNodeID, "job.start_server", map[string]any{
			"pair_id":    pair.ID,
			"protocol":   run.Protocol,
			"port":       port,
			"duration":   duration,
			"request_no": i + 1,
		}, 6*time.Second)
		if err == nil {
			if ok, _ := resp.Payload["ok"].(bool); ok {
				serverReady = true
				break
			}
		}
		time.Sleep(2 * time.Second)
	}
	if !serverReady {
		o.failPair(pair.ID, "iperf server not ready after retries")
		return
	}

	resp, err := o.hub.SendCommand(pair.SourceNodeID, "job.run_client", map[string]any{
		"pair_id":          pair.ID,
		"target_ip":        targetIP,
		"protocol":         run.Protocol,
		"port":             port,
		"duration":         duration,
		"parallel_streams": parallel,
		"udp_bandwidth":    udpBw,
	}, time.Duration(duration+10)*time.Second)
	if err != nil {
		o.failPair(pair.ID, err.Error())
		return
	}
	ok, _ := resp.Payload["ok"].(bool)
	if !ok {
		errMsg, _ := resp.Payload["error"].(string)
		if errMsg == "" {
			errMsg = "client execution failed"
		}
		o.failPair(pair.ID, errMsg)
		return
	}

	metrics, _ := resp.Payload["metrics"].(map[string]any)
	rawOutput, _ := resp.Payload["raw_output"].(string)
	if metrics == nil {
		metrics = map[string]any{}
	}
	if err := o.store.InsertResult(db.NewID("res"), pair.ID, run.Protocol, metrics, rawOutput); err != nil {
		o.failPair(pair.ID, "failed to save result")
		return
	}
	_ = o.store.UpdatePairStatus(pair.ID, "success", "", false, true)
}

func (o *Orchestrator) failPair(pairID, errMsg string) {
	_ = o.store.UpdatePairStatus(pairID, "failed", errMsg, false, true)
}

func (o *Orchestrator) acquire(a, b string) func() {
	o.globalSem <- struct{}{}
	ids := []string{a}
	if b != a {
		ids = append(ids, b)
	}
	sort.Strings(ids)

	for _, id := range ids {
		sem := o.nodeSemaphore(id)
		sem <- struct{}{}
	}
	return func() {
		for _, id := range ids {
			sem := o.nodeSemaphore(id)
			select {
			case <-sem:
			default:
			}
		}
		select {
		case <-o.globalSem:
		default:
		}
	}
}

func (o *Orchestrator) nodeSemaphore(nodeID string) chan struct{} {
	o.nodeMu.Lock()
	defer o.nodeMu.Unlock()
	sem, ok := o.nodeSem[nodeID]
	if !ok {
		sem = make(chan struct{}, 3)
		o.nodeSem[nodeID] = sem
	}
	return sem
}

func intFromParams(params map[string]any, key string, def int) int {
	v, ok := params[key]
	if !ok {
		return def
	}
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case string:
		parsed, err := strconv.Atoi(n)
		if err != nil {
			return def
		}
		return parsed
	default:
		return def
	}
}

func strFromParams(params map[string]any, key, def string) string {
	v, ok := params[key]
	if !ok {
		return def
	}
	s, ok := v.(string)
	if !ok || s == "" {
		return def
	}
	return s
}

func ValidatePairs(pairs []db.PairSpec) error {
	if len(pairs) == 0 {
		return errors.New("pairs can not be empty")
	}
	seen := map[string]struct{}{}
	for _, p := range pairs {
		if p.SourceNodeID == "" || p.TargetNodeID == "" {
			return errors.New("source_node_id and target_node_id are required")
		}
		key := p.SourceNodeID + "->" + p.TargetNodeID
		if _, ok := seen[key]; ok {
			return errors.New("duplicate pair: " + key)
		}
		seen[key] = struct{}{}
	}
	return nil
}
