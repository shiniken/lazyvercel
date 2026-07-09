package vercel

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type flexibleInt int64

func (f *flexibleInt) UnmarshalJSON(data []byte) error {
	if strings.TrimSpace(string(data)) == "null" {
		*f = 0
		return nil
	}

	var number int64
	if err := json.Unmarshal(data, &number); err == nil {
		*f = flexibleInt(number)
		return nil
	}

	var text string
	if err := json.Unmarshal(data, &text); err != nil {
		return err
	}
	if text == "" {
		*f = 0
		return nil
	}
	parsed, err := strconv.ParseInt(text, 10, 64)
	if err != nil {
		return fmt.Errorf("parse flexible int %q: %w", text, err)
	}
	*f = flexibleInt(parsed)
	return nil
}

func (f flexibleInt) Int64() int64 {
	return int64(f)
}

type flexibleText string

func (f *flexibleText) UnmarshalJSON(data []byte) error {
	var text string
	if err := json.Unmarshal(data, &text); err == nil {
		*f = flexibleText(text)
		return nil
	}

	var number int64
	if err := json.Unmarshal(data, &number); err == nil {
		*f = flexibleText(strconv.FormatInt(number, 10))
		return nil
	}

	var floating float64
	if err := json.Unmarshal(data, &floating); err == nil {
		*f = flexibleText(strconv.FormatFloat(floating, 'f', -1, 64))
		return nil
	}

	return nil
}

func (f flexibleText) String() string {
	return string(f)
}
