package request

type ProxyRequest struct {
	APIKey     string `json:"api_key"`
	Target     string `json:"target"`
	TargetPort string `json:"target_port"`
}

type ProxyResponse struct {
	Ok       bool   `json:"ok"`
	Err      string `json:"err"`
	BindAddr string `json:"bind_addr"`
	BindPort int `json:"bind_port"`
}
