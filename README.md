## Outputs

### Export-Id

- Desc: id of created export
- Variable-Name: export_id

## Camunda-Input-Variables

### Module-Data

- Desc: sets fields for Module.ModuleData
- Variable-Name-Template: `{{config.WorkerParamPrefix}}.module_data`
- Variable-Name-Example: `export.module_data`
- Value: `json.Marshal(map[string]interface{})`

### Request

- Desc: request forwarded to service-service to create export
- Variable-Name-Template: `{{config.WorkerParamPrefix}}.request`
- Variable-Name-Example: `export.request`
- Value: json.Marshal(ServingRequest{})
