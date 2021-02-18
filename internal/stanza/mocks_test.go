// Copyright The OpenTelemetry Authors
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

package stanza

import (
	"context"
	"fmt"

	"github.com/open-telemetry/opentelemetry-log-collection/entry"
	"github.com/open-telemetry/opentelemetry-log-collection/operator"
	"github.com/open-telemetry/opentelemetry-log-collection/operator/builtin/transformer/noop"
	"github.com/open-telemetry/opentelemetry-log-collection/operator/helper"
	"go.opentelemetry.io/collector/config/configmodels"
	"go.opentelemetry.io/collector/consumer/pdata"
)

// This file implements some useful testing components
func init() {
	operator.Register("unstartable_operator", func() operator.Builder { return NewUnstartableConfig() })
}

// UnstartableConfig is the configuration of an unstartable mock operator
type UnstartableConfig struct {
	helper.OutputConfig `yaml:",inline"`
}

// UnstartableOperator is an operator that will build but not start
// While this is not expected behavior, it is possible that build-time
// validation could be invalidated before Start() is called
type UnstartableOperator struct {
	helper.OutputOperator
}

func newUnstartableParams() map[string]interface{} {
	return map[string]interface{}{"type": "unstartable_operator"}
}

// NewUnstartableConfig creates new output config
func NewUnstartableConfig() *UnstartableConfig {
	return &UnstartableConfig{
		OutputConfig: helper.NewOutputConfig("unstartable_operator", "unstartable_operator"),
	}
}

// Build will build an unstartable operator
func (c *UnstartableConfig) Build(context operator.BuildContext) ([]operator.Operator, error) {
	o, _ := c.OutputConfig.Build(context)
	return []operator.Operator{&UnstartableOperator{OutputOperator: o}}, nil
}

// Start will return an error
func (o *UnstartableOperator) Start() error {
	return fmt.Errorf("something very unusual happened")
}

// Process will return nil
func (o *UnstartableOperator) Process(ctx context.Context, entry *entry.Entry) error {
	return nil
}

type mockLogsConsumer struct {
	received int
}

func (m *mockLogsConsumer) ConsumeLogs(ctx context.Context, ld pdata.Logs) error {
	m.received++
	return nil
}

type mockLogsRejecter struct {
	rejected int
}

func (m *mockLogsRejecter) ConsumeLogs(ctx context.Context, ld pdata.Logs) error {
	m.rejected++
	return fmt.Errorf("no")
}

const testType = "test"

type TestConfig struct {
	BaseConfig `mapstructure:",squash"`
	Input      InputConfig `mapstructure:",remain"`
}
type TestReceiverType struct{}

func (f TestReceiverType) Type() configmodels.Type {
	return configmodels.Type(testType)
}

func (f TestReceiverType) CreateDefaultConfig() configmodels.Receiver {
	return &TestConfig{
		BaseConfig: BaseConfig{
			ReceiverSettings: configmodels.ReceiverSettings{
				TypeVal: configmodels.Type(testType),
				NameVal: testType,
			},
			Operators: OperatorConfigs{},
		},
		Input: InputConfig{},
	}
}

func (f TestReceiverType) BaseConfig(cfg configmodels.Receiver) BaseConfig {
	return cfg.(*TestConfig).BaseConfig
}

func (f TestReceiverType) DecodeInputConfig(cfg configmodels.Receiver) (*operator.Config, error) {
	testConfig := cfg.(*TestConfig)

	// Allow tests to run without implementing input config
	if testConfig.Input["type"] == nil {
		return &operator.Config{Builder: noop.NewNoopOperatorConfig("nop")}, nil
	}

	// Allow tests to explicitly prompt a failure
	if testConfig.Input["type"] == "unknown" {
		return nil, fmt.Errorf("Unknown input type")
	}
	return &operator.Config{Builder: NewUnstartableConfig()}, nil
}