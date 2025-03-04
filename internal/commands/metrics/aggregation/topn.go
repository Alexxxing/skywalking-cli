// Licensed to Apache Software Foundation (ASF) under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Apache Software Foundation (ASF) licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package aggregation

import (
	"fmt"
	"strconv"

	api "skywalking.apache.org/repo/goapi/query"

	"github.com/urfave/cli/v2"

	"github.com/apache/skywalking-cli/internal/commands/interceptor"
	"github.com/apache/skywalking-cli/internal/flags"
	"github.com/apache/skywalking-cli/internal/logger"
	"github.com/apache/skywalking-cli/internal/model"
	"github.com/apache/skywalking-cli/pkg/display"
	"github.com/apache/skywalking-cli/pkg/display/displayable"
	"github.com/apache/skywalking-cli/pkg/graphql/metrics"
	"github.com/apache/skywalking-cli/pkg/graphql/utils"
)

var TopN = &cli.Command{
	Name:      "top",
	Usage:     "query the top <n> entities sorted by the specified metrics",
	ArgsUsage: "<n>",
	UsageText: `Query the top <n> entities sorted by the specified metrics.

Examples:
1. Query the top 5 services whose sla are largest:
$ swctl metrics top --name service_sla 5

2. Query the top 5 endpoints whose sla is largest:
$ swctl metrics top --name endpoint_sla 5

3. Query the top 5 instances of service "boutique::adservice" whose sla are largest:
$ swctl metrics top --name service_instance_sla --service-name boutique::adservice 5
`,
	Flags: flags.Flags(
		flags.DurationFlags,
		flags.MetricsFlags,
		flags.ServiceFlags,
		[]cli.Flag{
			&cli.GenericFlag{
				Name:  "order",
				Usage: "the `order` by which the top entities are sorted",
				Value: &model.OrderEnumValue{
					Enum:     api.AllOrder,
					Default:  api.OrderDes,
					Selected: api.OrderDes,
				},
			},
		},
	),
	Before: interceptor.BeforeChain(
		interceptor.DurationInterceptor,
		interceptor.ParseService(false),
	),
	Action: func(ctx *cli.Context) error {
		start := ctx.String("start")
		end := ctx.String("end")
		step := ctx.Generic("step").(*model.StepEnumValue).Selected

		metricsName := ctx.String("name")
		scope := utils.ParseScopeInTop(metricsName)
		order := ctx.Generic("order").(*model.OrderEnumValue).Selected
		topN := 5
		parentServiceID := ctx.String("service-id")
		parentService, normal, err := interceptor.ParseServiceID(parentServiceID)
		if err != nil {
			return err
		}

		if ctx.NArg() > 0 {
			nn, err2 := strconv.Atoi(ctx.Args().First())
			if err2 != nil {
				return fmt.Errorf("the 1st argument must be a number: %v", err2)
			}
			topN = nn
		}

		duration := api.Duration{
			Start: start,
			End:   end,
			Step:  step,
		}

		logger.Log.Debugln(metricsName, scope, topN)

		metricsValues, err := metrics.SortMetrics(ctx, api.TopNCondition{
			Name:          metricsName,
			ParentService: &parentService,
			Normal:        &normal,
			Scope:         &scope,
			TopN:          topN,
			Order:         order,
		}, duration)

		if err != nil {
			return err
		}

		return display.Display(ctx, &displayable.Displayable{Data: metricsValues})
	},
}
