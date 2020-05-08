// Copyright 2020 The PipeCD Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package executor

import (
	"fmt"
	"sync"

	"github.com/kapetaniosci/pipe/pkg/model"
)

type Registry interface {
	Register(stage model.Stage, f Factory) error
	Executor(stage model.Stage, in Input) (Executor, error)
}

type registry struct {
	factories map[model.Stage]Factory
	mu        sync.RWMutex
}

func (r *registry) Register(stage model.Stage, f Factory) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.factories[stage]; ok {
		return fmt.Errorf("executor for %s stage has already registered", stage)
	}
	r.factories[stage] = f
	return nil
}

func (r *registry) Executor(stage model.Stage, in Input) (Executor, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	f, ok := r.factories[stage]
	if !ok {
		return nil, fmt.Errorf("no registered executor for stage %s", stage)
	}
	return f(in), nil
}

var defaultRegistry = &registry{}

func DefaultRegistry() Registry {
	return defaultRegistry
}