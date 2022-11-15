package main

import (
	"fmt"
	"github.com/addit-digital/addcache"
	"time"
)

func main() {
	// NewCache method creates cache with default values
	cache := addcache.NewCache()

	// Cleaning of memory can be stopped manually
	defer cache.StopCleanup()

	// CreateKey creates key with semicolon delimited
	userKey := cache.CreateKey("user", "12")
	userData := User{
		Id:       2,
		Name:     "Test",
		Lastname: "Test",
	}
	// Persisting data into cache
	cache.Set(userKey, userData)

	// Getting data from cache with key
	user, err := cache.Get(userKey)
	if err != nil {
		fmt.Errorf("cache error - %v", err)
	}
	fmt.Print(user.(User))

	// Deleting cached data manually
	cache.Delete(userKey)

	// Persist data with expiration
	cache.SetEx(userKey, userData, 40*time.Second)
}

type User struct {
	Id       int64
	Name     string
	Lastname string
}
