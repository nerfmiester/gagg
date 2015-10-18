package Structs

type ProvisionReq struct {
	source 	string 	`json:"sourceEnvironment"`
	target 	string 	`json:"targetEnvironment"`
	channel	string	`json:"channel"`
	action	string	`json:"action"`
}

type FlightsReq struct {
	origin 		string
	destination string
	startDate	string
	endDate		string
}
