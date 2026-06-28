package types

type FieldType string

const (
	FieldString  FieldType = "string"
	FieldInteger FieldType = "integer"
	FieldNumber  FieldType = "number"
	FieldBoolean FieldType = "boolean"
	FieldObject  FieldType = "object"
	FieldArray   FieldType = "array"
)

type FieldSchema struct {
	Type        FieldType `json:"type"`
	Required    bool      `json:"required,omitempty"`
	Description string    `json:"description,omitempty"`
	Default     any       `json:"default,omitempty"`
}

type Schema struct {
	Fields              map[string]FieldSchema `json:"fields,omitempty"`
	RejectUnknownFields bool                   `json:"reject_unknown_fields"`
}
