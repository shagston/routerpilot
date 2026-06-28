package safety

import (
	"context"
	"fmt"

	"github.com/shagston/routerpilot/sdk/tool"
	"github.com/shagston/routerpilot/sdk/types"
)

type Config struct {
	Permissions  []types.Permission
	Capabilities []types.Capability
}

type Validator struct {
	registry     tool.Registry
	permissions  map[types.Permission]bool
	capabilities map[types.Capability]bool
}

func NewValidator(registry tool.Registry, config Config) *Validator {
	validator := &Validator{
		registry:     registry,
		permissions:  make(map[types.Permission]bool, len(config.Permissions)),
		capabilities: make(map[types.Capability]bool, len(config.Capabilities)),
	}
	for _, permission := range config.Permissions {
		validator.permissions[permission] = true
	}
	for _, capability := range config.Capabilities {
		validator.capabilities[capability] = true
	}
	return validator
}

func (v *Validator) Validate(ctx context.Context, plan types.Plan) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	for _, task := range plan.Steps {
		t, err := v.registry.Get(task.Tool)
		if err != nil {
			return err
		}
		metadata := t.Metadata()
		if err := v.validatePermissions(task.Tool, metadata.Permissions); err != nil {
			return err
		}
		if err := v.validateCapabilities(task.Tool, metadata.Capabilities); err != nil {
			return err
		}
		if err := ValidateInput(t.InputSchema(), task.Arguments); err != nil {
			return fmt.Errorf("%w: task %s", err, task.ID)
		}
	}

	return nil
}

func (v *Validator) validatePermissions(toolID types.ToolID, permissions []types.Permission) error {
	for _, permission := range permissions {
		if !v.permissions[permission] {
			return fmt.Errorf("%w: tool %s requires %s", types.ErrPermissionDenied, toolID, permission)
		}
	}
	return nil
}

func (v *Validator) validateCapabilities(toolID types.ToolID, capabilities []types.Capability) error {
	for _, capability := range capabilities {
		if !v.capabilities[capability] {
			return fmt.Errorf("%w: tool %s requires %s", types.ErrCapabilityMissing, toolID, capability)
		}
	}
	return nil
}

func ValidateInput(schema types.Schema, input types.ToolInput) error {
	if input == nil {
		input = types.ToolInput{}
	}

	for name, field := range schema.Fields {
		value, exists := input[name]
		if field.Required && !exists {
			return fmt.Errorf("%w: missing required field %s", types.ErrInvalidInput, name)
		}
		if !exists {
			continue
		}
		if !matchesType(field.Type, value) {
			return fmt.Errorf("%w: field %s must be %s", types.ErrInvalidInput, name, field.Type)
		}
	}

	if schema.RejectUnknownFields {
		for name := range input {
			if _, exists := schema.Fields[name]; !exists {
				return fmt.Errorf("%w: unknown field %s", types.ErrInvalidInput, name)
			}
		}
	}

	return nil
}

func matchesType(fieldType types.FieldType, value any) bool {
	switch fieldType {
	case types.FieldString:
		_, ok := value.(string)
		return ok
	case types.FieldInteger:
		switch v := value.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			return true
		case float64:
			return v == float64(int64(v))
		case float32:
			return float64(v) == float64(int64(v))
		default:
			return false
		}
	case types.FieldNumber:
		switch value.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
			return true
		default:
			return false
		}
	case types.FieldBoolean:
		_, ok := value.(bool)
		return ok
	case types.FieldObject:
		_, ok := value.(map[string]any)
		return ok
	case types.FieldArray:
		switch value.(type) {
		case []any, []string, []int:
			return true
		default:
			return false
		}
	default:
		return false
	}
}
