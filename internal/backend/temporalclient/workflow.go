package temporalclient

import (
	"time"

	"go.temporal.io/sdk/client"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
)

var defaultWorkflowConfig = WorkflowConfig{}

// Common way to define configuration that can be used in multiple modules,
// saving the need to repeat the same configuration in each module.
type WorkflowConfig struct {
	WorkflowTaskTimeout time.Duration `koanf:"workflow_task_timeout"`
}

// other overrides self.
func (wc WorkflowConfig) With(other WorkflowConfig) WorkflowConfig {
	return WorkflowConfig{
		WorkflowTaskTimeout: kittehs.FirstNonZero(other.WorkflowTaskTimeout, wc.WorkflowTaskTimeout),
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
