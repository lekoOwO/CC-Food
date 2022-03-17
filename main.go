package main

func main() {
	db := getDB()
	r := initAPI(db)
	r.Run()
}
