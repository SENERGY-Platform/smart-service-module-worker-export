/*
 * Copyright (c) 2022 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package pkg

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/SENERGY-Platform/smart-service-module-worker-export/pkg/export"
	lib "github.com/SENERGY-Platform/smart-service-module-worker-lib"
	"github.com/SENERGY-Platform/smart-service-module-worker-lib/pkg/auth"
	"github.com/SENERGY-Platform/smart-service-module-worker-lib/pkg/camunda"
	"github.com/SENERGY-Platform/smart-service-module-worker-lib/pkg/configuration"
	"github.com/SENERGY-Platform/smart-service-module-worker-lib/pkg/model"
	"github.com/SENERGY-Platform/smart-service-module-worker-lib/pkg/smartservicerepository"
)

func Start(ctx context.Context, wg *sync.WaitGroup, config export.Config, libConfig configuration.Config) error {
	handlerFactory := func(auth *auth.Auth, smartServiceRepo *smartservicerepository.SmartServiceRepository) (camunda.Handler, error) {
		handler := export.New(
			config,
			libConfig,
			auth,
			smartServiceRepo,
		)

		interval, err := time.ParseDuration(config.HealthCheckInterval)
		if err != nil {
			return nil, err
		}

		healthCheck := func(ctx context.Context, module model.SmartServiceModule) (health error, err error) {
			token, err := auth.ExchangeUserToken(module.UserId)
			if err != nil {
				return nil, err
			}
			exportId, err := getExportId(module.ModuleData)
			if err != nil {
				return nil, err
			}
			code, err := handler.CheckExport(ctx, token, exportId)
			if err != nil {
				return nil, err
			}
			if code >= 300 {
				return fmt.Errorf("export health check returned status-code %v", code), nil
			}
			return nil, nil
		}
		moduleQuery := model.ModulQuery{TypeFilter: &libConfig.CamundaWorkerTopic}
		smartServiceRepo.StartHealthCheck(ctx, interval, moduleQuery, healthCheck) //timer loop
		smartServiceRepo.RunHealthCheck(ctx, moduleQuery, healthCheck)             //initial check
		return handler, nil
	}
	return lib.Start(ctx, wg, libConfig, handlerFactory)
}

func getExportId(moduleData map[string]interface{}) (string, error) {
	exp, ok := moduleData["export"]
	if !ok {
		return "", fmt.Errorf("missing export in module data")
	}
	exportObj, ok := exp.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid export in module data")
	}
	exportId, ok := exportObj["ID"].(string)
	if !ok {
		return "", fmt.Errorf("invalid export in module data (id is not string)")
	}
	return exportId, nil
}
