package redkeep

import "strings"

//GetValue works like this:
//from must be a selector like user.comment.author
//GetValue then looks recursively for that element
//therefore all of the following return values are possible
//map[string]interface{}
//nil
//string
//or basic mongodb types
func GetValue(from string, ds interface{}) interface{} {
	data, ok := ds.(map[string]interface{})
	if !ok {
		return nil
	}

	if index := strings.Index(from, "."); index != -1 {
		return GetValue(from[index+1:], data[from[:index]])
	}

	return data[from]
}

//HasKey will return wether the key was found or not
func HasKey(key string, ds interface{}) bool {
	data, ok := ds.(map[string]interface{})
	if !ok {
		return false
	}

	if index := strings.Index(key, "."); index != -1 {
		return HasKey(key[index+1:], data[key[:index]])
	}

	if _, ok := data[key]; ok {
		return true
	}

	return false
}
