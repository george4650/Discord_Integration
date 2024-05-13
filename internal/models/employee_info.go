package models

type EmployeeInfo struct {
	Data []struct {
		EmployeeId    int    `json:"employee_id"`
		EmployeeName  string `json:"employee_name"`
		MainTagId     int    `json:"main_tag_id"`
		TagName       string `json:"tag_name"`
		SecondaryTags []struct {
			TagId   int    `json:"tag_id"`
			TagName string `json:"tag_name"`
		} `json:"secondary_tags"`
	} `json:"data"`
}
