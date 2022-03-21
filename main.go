package main

func main() {
	db := getDB()
	r := initAPI(db)
	r.Run("172.21.192.1:8080")
}
