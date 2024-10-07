package robin

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

type Comment struct {
	JobDataFilename string `json:"job_data_filename"`
}

func (c *Comment) IsEmpty() bool {
	return c.JobDataFilename == ""
}

func (c *Comment) Encode() string {
	commentJSON, _ := json.Marshal(c)
	return base64.RawStdEncoding.EncodeToString(commentJSON)
}

func (c *Comment) Decode(data string) error {
	if data == "" {
		return nil
	}
	commentJSON, err := base64.RawStdEncoding.DecodeString(data)
	if err != nil {
		return fmt.Errorf("decode base64: %w", err)
	}
	err = json.Unmarshal(commentJSON, c)
	if err != nil {
		return fmt.Errorf("unmarshal JSON: %w", err)
	}
	return nil
}
