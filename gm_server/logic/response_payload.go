package logic

import "encoding/json"

type forwardedResponse struct {
	ErrCode int             `json:"errcode"`
	ErrMsg  string          `json:"errmsg"`
	Data    json.RawMessage `json:"data"`
}

func decodeForwardedData(body string, out any) error {
	var resp forwardedResponse
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		return err
	}
	if len(resp.Data) == 0 || string(resp.Data) == "null" {
		return nil
	}
	return json.Unmarshal(resp.Data, out)
}
