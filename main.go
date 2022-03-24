package main

import "os"

func main() {
	db := getDB()
	r := initAPI(db)
	r.Run(os.Getenv("CC_FOOD_API"))
}
