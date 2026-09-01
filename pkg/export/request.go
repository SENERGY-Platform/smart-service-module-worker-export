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
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"runtime/debug"
	"time"

	"github.com/SENERGY-Platform/gin-middleware/otelx"
	"github.com/SENERGY-Platform/smart-service-module-worker-lib/pkg/auth"
)

func (this *Export) send(ctx context.Context, token auth.Token, request ServingRequest) (result Instance, err error) {
	body, err := json.Marshal(request)
	if err != nil {
		return result, err
	}
	this.libConfig.GetLogger().DebugContext(ctx, "send export request", "request", string(body))
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
	err = otelx.InjectContextToRequest(ctx, req)
	if err != nil {
		return result, err
	}
	req.Header.Set("Authorization", token.Jwt())
	req.Header.Set("X-UserId", token.GetUserId())
	this.libConfig.GetLogger().DebugContext(ctx, "send export request", "request", string(body))
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

var DefaultTimeout = 30 * time.Second

func (this *Export) CheckExport(ctx context.Context, token auth.Token, id string) (code int, err error) {
	client := http.Client{
		Timeout: DefaultTimeout,
	}
	req, err := http.NewRequest(
		"GET",
		this.config.ServingServiceUrl+"/instance/"+url.PathEscape(id),
		nil,
	)
	if err != nil {
		this.libConfig.GetLogger().ErrorContext(ctx, "error in CheckExport", "error", err, "stack", string(debug.Stack()))
		return 0, err
	}
	err = otelx.InjectContextToRequest(ctx, req)
	if err != nil {
		this.libConfig.GetLogger().ErrorContext(ctx, "error in CheckExport", "error", err, "stack", string(debug.Stack()))
		return 0, err
	}
	req.Header.Set("Authorization", token.Jwt())
	req.Header.Set("X-UserId", token.GetUserId())

	this.libConfig.GetLogger().DebugContext(ctx, "check export request", "url", req.URL.String(), "method", req.Method, "xuser", req.Header.Get("X-UserId"))

	resp, err := client.Do(req)
	if err != nil {
		this.libConfig.GetLogger().ErrorContext(ctx, "error in CheckExport", "error", err, "stack", string(debug.Stack()))
		return 0, err
	}
	defer resp.Body.Close()
	return resp.StatusCode, nil
}
