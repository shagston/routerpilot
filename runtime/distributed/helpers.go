package distributed

import "encoding/json"

func parseCapabilityRequest(payload []byte) (capabilityRequest, error) {
	var req capabilityRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		return req, err
	}
	return req, nil
}

func serializePayload(v any) ([]byte, error) {
	return json.Marshal(v)
}
