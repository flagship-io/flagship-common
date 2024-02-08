package decision

import (
	"sync"
	"time"
)

type assignmentResult struct {
	result *VisitorAssignments
	idType string
	err    error
}

type allVisitorAssignments struct {
	Standard      *VisitorAssignments
	Anonymous     *VisitorAssignments
	DecisionGroup *VisitorAssignments
}

func getCache(
	environmentID string,
	visitorID string,
	anonymousID string,
	decisionGroup string,
	enableReconciliation bool,
	getCacheHandler func(environmentID string, id string) (*VisitorAssignments, error)) (*allVisitorAssignments, error) {

	cacheChan := make(chan (*assignmentResult))
	allAssignments := &allVisitorAssignments{
		Standard: &VisitorAssignments{
			Assignments: map[string]*VisitorCache{},
		},
	}

	var err error
	var nbRoutines = 1

	fetchCacheForID := func(c chan (*assignmentResult), id string, idType string) {
		logger.Logf(InfoLevel, "getting assignment cache for %s: %s", idType, id)
		newAssignments, err := getCacheHandler(environmentID, id)
		c <- &assignmentResult{
			result: newAssignments,
			idType: idType,
			err:    err,
		}
	}

	go fetchCacheForID(cacheChan, visitorID, "standard")
	if enableReconciliation {
		nbRoutines++
		go fetchCacheForID(cacheChan, anonymousID, "anonymous")
	}
	if decisionGroup != "" {
		nbRoutines++
		go fetchCacheForID(cacheChan, decisionGroup, "decisionGroup")
	}

	for i := 0; i < nbRoutines; i++ {
		r := <-cacheChan
		switch r.idType {
		case "standard":
			allAssignments.Standard = r.result
		case "anonymous":
			allAssignments.Anonymous = r.result
		case "decisionGroup":
			allAssignments.DecisionGroup = r.result
		}
		err = r.err
	}

	return allAssignments, err
}

// Saves a set of cache assignments for a specific id type and using cache handlers
func saveCacheAssignments(
	wg *sync.WaitGroup,
	handlers DecisionHandlers,
	envID string,
	id string,
	idType string,
	assignments map[string]*VisitorCache,
) {
	if len(assignments) == 0 || id == "" {
		return
	}

	wg.Add(1)
	now := time.Now()
	go func() {
		defer wg.Done()
		logger.Logf(InfoLevel, "saving assignments cache for %s: %s", idType, id)
		err := handlers.SaveCache(envID, id, &VisitorAssignments{
			Timestamp:   now.Unix(),
			Assignments: assignments,
		})
		if err != nil {
			logger.Logf(ErrorLevel, "error occurred on cache saving for %s: %v", id, err)
		}
	}()
}
