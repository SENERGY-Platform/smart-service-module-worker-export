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

package export

import (
	"fmt"
	"github.com/SENERGY-Platform/smart-service-module-worker-lib/pkg/auth"
	"github.com/SENERGY-Platform/smart-service-module-worker-lib/pkg/configuration"
	"github.com/SENERGY-Platform/smart-service-module-worker-lib/pkg/model"
	"github.com/go-playground/validator/v10"
	"io"
	"log"
	"net/http"
	"net/url"
	"runtime/debug"
)

func New(config Config, libConfig configuration.Config, auth *auth.Auth, smartServiceRepo SmartServiceRepo) *Export {
	return &Export{config: config, libConfig: libConfig, auth: auth, smartServiceRepo: smartServiceRepo}
}

type Export struct {
	config           Config
	libConfig        configuration.Config
	auth             *auth.Auth
	smartServiceRepo SmartServiceRepo
}

type SmartServiceRepo interface {
	GetInstanceUser(instanceId string) (userId string, err error)
}

func (this *Export) Do(task model.CamundaExternalTask) (modules []model.Module, outputs map[string]interface{}, err error) {
	userId, err := this.smartServiceRepo.GetInstanceUser(task.ProcessInstanceId)
	if err != nil {
		log.Println("ERROR: unable to get instance user", err)
		return modules, outputs, err
	}
	token, err := this.auth.ExchangeUserToken(userId)
	if err != nil {
		log.Println("ERROR: unable to exchange user token", err)
		return modules, outputs, err
	}

	defaultModuleData, analyticsModuleDeleteInfo, outputs, err := this.do(token, task)
	if err != nil {
		return modules, outputs, err
	}
	moduleData := this.getModuleData(task)
	for key, value := range defaultModuleData {
		moduleData[key] = value
	}

	return []model.Module{{
			Id:               this.getModuleId(task),
			ProcesInstanceId: task.ProcessInstanceId,
			SmartServiceModuleInit: model.SmartServiceModuleInit{
				DeleteInfo: analyticsModuleDeleteInfo,
				ModuleType: this.libConfig.CamundaWorkerTopic,
				ModuleData: moduleData,
			},
		}},
		outputs,
		err
}

func (this *Export) Undo(modules []model.Module, reason error) {
	log.Println("UNDO:", reason)
	for _, module := range modules {
		if module.DeleteInfo != nil {
			err := this.useModuleDeleteInfo(*module.DeleteInfo)
			if err != nil {
				log.Println("ERROR:", err)
				debug.PrintStack()
			}
		}
	}
}

func (this *Export) useModuleDeleteInfo(info model.ModuleDeleteInfo) error {
	req, err := http.NewRequest("DELETE", info.Url, nil)
	if err != nil {
		return err
	}
	if info.UserId != "" {
		token, err := this.auth.ExchangeUserToken(info.UserId)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", token.Jwt())
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 && resp.StatusCode != http.StatusNotFound {
		temp, _ := io.ReadAll(resp.Body)
		err = fmt.Errorf("unexpected response: %v, %v", resp.StatusCode, string(temp))
		log.Println("ERROR:", err)
		debug.PrintStack()
		return err
	}
	_, _ = io.ReadAll(resp.Body)
	return nil
}

func (this *Export) getModuleId(task model.CamundaExternalTask) string {
	return task.ProcessInstanceId + "." + task.Id
}

func (this *Export) do(token auth.Token, task model.CamundaExternalTask) (moduleData map[string]interface{}, deleteInfo *model.ModuleDeleteInfo, outputs map[string]interface{}, err error) {
	request, err := this.getRequest(task)
	if err != nil {
		return moduleData, deleteInfo, outputs, err
	}
	transformedRequest, err := this.validateAndTransformRequest(token, task, request)
	if err != nil {
		return moduleData, deleteInfo, outputs, fmt.Errorf("invalid export request: %w", err)
	}

	exportInstance, err := this.send(token, transformedRequest)
	if err != nil {
		return moduleData, deleteInfo, outputs, err
	}

	return map[string]interface{}{
			"export": exportInstance,
		}, &model.ModuleDeleteInfo{
			Url:    this.config.ServingServiceUrl + "/instance/" + url.PathEscape(exportInstance.ID.String()),
			UserId: token.GetUserId(),
		}, map[string]interface{}{
			"export_id": exportInstance.ID.String(),
		}, nil
}

func (this *Export) validateAndTransformRequest(token auth.Token, task model.CamundaExternalTask, request ServingRequest) (result ServingRequest, err error) {
	err = validator.New().Struct(request)
	result = request
	return result, err
}
