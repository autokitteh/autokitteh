package temporalclient

import (
	"cmp"
	"time"

	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/workflow"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
)

var defaultWorkflowConfig = WorkflowConfig{
	WorkflowTaskTimeout: 60 * time.Second,
}

// Common way to define configuration that can be used in multiple modules,
// saving the need to repeat the same configuration in each module.
type WorkflowConfig struct {
	WorkflowTaskTimeout time.Duration `koanf:"workflow_task_timeout"`
}

// other overrides self.
func (wc WorkflowConfig) With(other WorkflowConfig) WorkflowConfig {
	return WorkflowConfig{
		WorkflowTaskTimeout: cmp.Or(other.WorkflowTaskTimeout, wc.WorkflowTaskTimeout),
	}
}

func (wc WorkflowConfig) ToStartWorkflowOptions(qname, id, sum string, memo map[string]string) client.StartWorkflowOptions {
	wc = wc.With(defaultWorkflowConfig)
	return client.StartWorkflowOptions{
		ID:                  id,
		TaskQueue:           qname,
		StaticSummary:       sum,
		Memo:                kittehs.TransformMapValues(memo, func(v string) any { return v }),
		WorkflowTaskTimeout: wc.WorkflowTaskTimeout,
	}
}

func (wc WorkflowConfig) ToChildWorkflowOptions(qname, id, sum string, pcp enumspb.ParentClosePolicy, memo map[string]string) workflow.ChildWorkflowOptions {
	wc = wc.With(defaultWorkflowConfig)
	return workflow.ChildWorkflowOptions{
		WorkflowID:          id,
		TaskQueue:           qname,
		StaticSummary:       sum,
		ParentClosePolicy:   pcp,
		Memo:                kittehs.TransformMapValues(memo, func(v string) any { return v }),
		WorkflowTaskTimeout: wc.WorkflowTaskTimeout,
	}
}
