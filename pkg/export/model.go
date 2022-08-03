/*
 * Copyright 2020 InfAI (CC SES)
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
	uuid "github.com/satori/go.uuid"
	"time"
)

const ( //DatabaseType
	InfluxDB    = "influxdb"
	TimescaleDB = "timescaledb"
)

const ( //FilterType
	TypeDevice    = "deviceId"   //not handled
	TypeAnalytics = "operatorId" // Filter = "{{pipelineId}}:{{operatorId}}"; Topic = "analytics-{{operatorName}}"; EntityName = "{{operatorId}}"; ServiceName = "{{operatorName}}"
	TypeImport    = "import_id"  // Filter = "{{import.id}}"; Topic = "{{import.kafka_topic}}"; EntityName = "{{import.name}}"; ServiceName = "{{import.import_type_id}}"
)

const ( //TimePrecision
	S  = "s"
	MS = "ms"
)

const ( //Offset
	Largest  = "largest"
	Smallest = "smallest"
)

//ServingRequest is a request to the serving-service to create a export (represented by Instance)
type ServingRequest struct {
	FilterType       string                `json:"FilterType,omitempty" validate:"required"`
	Filter           string                `json:"Filter,omitempty" validate:"required"`
	Name             string                `json:"Name,omitempty" validate:"required"`
	EntityName       string                `json:"EntityName,omitempty" validate:"required"`
	ServiceName      string                `json:"ServiceName,omitempty" validate:"required"`
	Description      string                `json:"Description,omitempty"`
	Topic            string                `json:"Topic,omitempty" validate:"required"`
	TimePath         string                `json:"TimePath,omitempty"`
	TimePrecision    string                `json:"TimePrecision,omitempty"`
	Generated        bool                  `json:"generated,omitempty"`
	Offset           string                `json:"Offset,omitempty" validate:"required"`
	ForceUpdate      bool                  `json:"ForceUpdate,omitempty"`
	Values           []ServingRequestValue `json:"Values,omitempty"`
	ExportDatabaseID string                `json:"ExportDatabaseID,omitempty"`
	TimestampFormat  string                `json:"TimestampFormat,omitempty"`
}

type ServingRequestValue struct {
	Name string `json:"Name,omitempty"`
	Type string `json:"Type,omitempty"`
	Path string `json:"Path,omitempty"`
	Tag  bool   `json:"Tag"`
}

//Instance is the response from serving-service representing an existing export
type Instance struct {
	ID               uuid.UUID      `gorm:"primary_key;type:char(36);column:id"`
	Name             string         `gorm:"type:varchar(255)"`
	Description      string         `gorm:"type:varchar(255)"`
	EntityName       string         `gorm:"type:varchar(255)"`
	ServiceName      string         `gorm:"type:varchar(255)"`
	Topic            string         `gorm:"type:varchar(255)"`
	ApplicationId    uuid.UUID      `gorm:"type:char(36)"`
	Database         string         `gorm:"type:varchar(255)"`
	Measurement      string         `gorm:"type:varchar(255)"`
	Filter           string         `gorm:"type:varchar(255)"`
	FilterType       string         `gorm:"type:varchar(255)"`
	TimePath         string         `gorm:"type:varchar(255)"`
	TimePrecision    *string        `gorm:"type:varchar(255)"`
	UserId           string         `gorm:"type:varchar(255)"`
	Generated        bool           `gorm:"type:bool;DEFAULT:false"`
	RancherServiceId string         `gorm:"type:varchar(255)"`
	Offset           string         `gorm:"type:varchar(255)"`
	ExportDatabaseID string         `gorm:"type:varchar(255)"`
	ExportDatabase   ExportDatabase `gorm:"association_autoupdate:false;association_autocreate:false"`
	TimestampFormat  string         `gorm:"type:varchar(255)"`
	Values           []Value        `gorm:"foreignkey:InstanceID;association_foreignkey:ID"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type Value struct {
	InstanceID uuid.UUID `gorm:"type:char(36)"`
	Name       string    `gorm:"type:varchar(255)"`
	Type       string    `gorm:"type:varchar(255)"`
	Path       string    `gorm:"type:varchar(255)"`
	Tag        bool      `gorm:"type:bool;DEFAULT:false"`
}

type ExportDatabase struct {
	ID            string `gorm:"primary_key;type:varchar(255);column:id"`
	Name          string `gorm:"type:varchar(255)"`
	Description   string `gorm:"type:varchar(255)"`
	Type          string `gorm:"type:varchar(255)"`
	Deployment    string `gorm:"type:varchar(255)"`
	Url           string `gorm:"type:varchar(255)"`
	EwFilterTopic string `gorm:"type:varchar(255)"`
	UserId        string `gorm:"type:varchar(255)"`
	Public        bool   `gorm:"type:bool;DEFAULT:false"`
}
