// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package components

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/client-go/util/workqueue"
)

// ComponentWalker is a worker pool that walks through component references and processes them using the provided item function.
type ComponentWalker struct {
	components     *Components
	waitGroup      sync.WaitGroup
	queue          workqueue.Queue[ComponentReference]
	lock           sync.Mutex
	log            logr.Logger
	workers        int
	itemsPreparing atomic.Int32
	itemFunc       ComponentReferenceFunc
	errs           []error
}

// ComponentReferenceFunc is a function that takes a ComponentReference and returns a slice of ComponentReferences to be processed next.
type ComponentReferenceFunc func(ComponentReference) ([]ComponentReference, error)

// NewComponentWalker creates a new ComponentWalker.
func NewComponentWalker(log logr.Logger, components *Components, workers int, itemFunc ComponentReferenceFunc) *ComponentWalker {
	return &ComponentWalker{
		components: components,
		workers:    workers,
		itemFunc:   itemFunc,
		queue:      workqueue.DefaultQueue[ComponentReference](),
		log:        log.WithName("component-walker"),
	}
}

// Start starts the worker goroutines to process the component references in the queue.
func (w *ComponentWalker) Start() {
	for i := 0; i < w.workers; i++ {
		w.waitGroup.Add(1)
		go w.worker()
	}
}

// Walk starts walking the components starting from the given root component reference.
func (w *ComponentWalker) Walk(root ComponentReference) error {
	w.pushComponentReference(root)
	w.Start()
	w.waitGroup.Wait()

	if len(w.errs) > 0 {
		return fmt.Errorf("errors occurred during walking components: %v", errors.Join(w.errs...))
	}
	return nil
}

func (w *ComponentWalker) worker() {
	for {
		w.itemsPreparing.Add(1)
		nextItem := w.popComponentReference()
		if nextItem == nil {
			w.itemsPreparing.Add(-1)
			if w.itemsPreparing.Load() == 0 {
				w.waitGroup.Done()
				return // No more items to process, exit the worker
			}
			time.Sleep(100 * time.Millisecond) // Wait before checking again
			continue
		}
		err := w.processComponentReference(*nextItem)
		w.itemsPreparing.Add(-1)
		if err != nil {
			w.addError(err, "failed to process component reference", *nextItem)
		}
	}
}

func (w *ComponentWalker) addError(err error, message string, item ComponentReference) {
	w.lock.Lock()
	defer w.lock.Unlock()
	w.errs = append(w.errs, fmt.Errorf("%s: %s: %w", message, item, err))
}

func (w *ComponentWalker) processComponentReference(item ComponentReference) error {
	newItems, err := w.itemFunc(item)
	if err != nil {
		return err
	}

	for _, newItem := range newItems {
		w.queue.Push(newItem)
	}

	return nil
}

func (w *ComponentWalker) pushComponentReference(cref ComponentReference) {
	w.lock.Lock()
	defer w.lock.Unlock()

	w.log.Info("Added component to queue", "component", cref)
	w.queue.Push(cref)
}

func (w *ComponentWalker) popComponentReference() *ComponentReference {
	w.lock.Lock()
	defer w.lock.Unlock()

	if w.queue.Len() == 0 {
		return nil
	}

	item := w.queue.Pop()
	return &item
}
