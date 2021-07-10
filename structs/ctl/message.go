package ctl

type Request struct {
	Command string `json:"command"`
	Text    string `json:"text"`
}

type Response struct {
	Status int
}
