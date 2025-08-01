/*
Copyright 2021 The Dapr Authors
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package command

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/camunda/zeebe/clients/go/v8/pkg/entities"

	"github.com/dapr/components-contrib/bindings"
	"github.com/dapr/kit/metadata"
)

var (
	ErrMissingJobType           = errors.New("jobType is a required attribute")
	ErrMissingMaxJobsToActivate = errors.New("maxJobsToActivate is a required attribute")
)

type activateJobsPayload struct {
	JobType           string             `json:"jobType"`
	MaxJobsToActivate *int32             `json:"maxJobsToActivate"`
	Timeout           *metadata.Duration `json:"timeout,omitempty"`
	WorkerName        string             `json:"workerName"`
	FetchVariables    []string           `json:"fetchVariables"`
	RequestTimeout    *metadata.Duration `json:"requestTimeout,omitempty"`
}

func (z *ZeebeCommand) activateJobs(ctx context.Context, req *bindings.InvokeRequest) (*bindings.InvokeResponse, error) {
	var payload activateJobsPayload
	err := json.Unmarshal(req.Data, &payload)
	if err != nil {
		return nil, err
	}

	if payload.JobType == "" {
		return nil, ErrMissingJobType
	}

	if payload.MaxJobsToActivate == nil {
		return nil, ErrMissingMaxJobsToActivate
	}

	cmd := z.client.NewActivateJobsCommand().
		JobType(payload.JobType).
		MaxJobsToActivate(*payload.MaxJobsToActivate)

	if payload.Timeout != nil && payload.Timeout.Duration != time.Duration(0) {
		cmd = cmd.Timeout(payload.Timeout.Duration)
	}

	if payload.WorkerName != "" {
		cmd = cmd.WorkerName(payload.WorkerName)
	}

	if payload.FetchVariables != nil {
		cmd = cmd.FetchVariables(payload.FetchVariables...)
	}

	var response []entities.Job
	if payload.RequestTimeout != nil && payload.RequestTimeout.Duration != time.Duration(0) {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, payload.RequestTimeout.Duration)
		defer cancel()
		response, err = cmd.Send(ctxWithTimeout)
	} else {
		response, err = cmd.Send(ctx)
	}

	if err != nil {
		return nil, fmt.Errorf("cannot activate jobs for type %s: %w", payload.JobType, err)
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal response to json: %w", err)
	}

	return &bindings.InvokeResponse{
		Data: jsonResponse,
	}, nil
}
