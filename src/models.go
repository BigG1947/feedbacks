package main

type Feedback struct {
	IdEmployee int    `json:"id_employee"`
	Mark       int    `json:"mark"`
	Text       string `json:"text"`
	TimeStamp  int64  `json:"time"`
}
