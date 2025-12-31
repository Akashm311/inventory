package main


func main() {
    app := App{}
    app.Initialise(DbUser, DbPassword, DbName)


    // Register HTTP routes.
    app.handleRoutes()


    // Start the server on localhost at port 10000.
    app.Run("localhost:10000")
}