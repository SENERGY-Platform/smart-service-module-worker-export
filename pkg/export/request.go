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
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/SENERGY-Platform/smart-service-module-worker-lib/pkg/auth"
)

func (this *Export) send(token auth.Token, request ServingRequest) (result Instance, err error) {
	body, err := json.Marshal(request)
	if err != nil {
		return result, err
	}
	this.libConfig.GetLogger().Debug("send export request", "request", string(body))
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	req, err := http.NewRequest(
		"POST",
		this.config.ServingServiceUrl+"/instance",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return result, err
	}
	req.Header.Set("Authorization", token.Jwt())
	req.Header.Set("X-UserId", token.GetUserId())
	this.libConfig.GetLogger().Debug("send export request", "request", string(body), "token", req.Header.Get("Authorization"))
	resp, err := client.Do(req)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return result, errors.New("unexpected statuscode")
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	return result, err
}
