package main

func main() {
	db := getDB()
	r := initAPI(db)
	r.Run("127.69.0.1:8080")
}
