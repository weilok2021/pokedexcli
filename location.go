package main

type locationList struct {
	Count int `json:"count"`
	Next string `json:"next"`
	Previous string `json:"previous"`
	Results []location `json:"results"`
}

type location struct {
	Name string `json:"name"`
	Url string `json:"url"`
}