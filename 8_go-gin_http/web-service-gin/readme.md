Create a module in which you can manage dependencies

Run the `go mod init` command, giving it the path of the module your code will be in.

```shell
$ go mod init example.com/web-service-gin
go: creating new go.mod: module example.com/web-service-gin
```



Begin tracking the Gin module as a dependency.

At the command line, use [`go get`](https://golang.org/cmd/go/#hdr-Add_dependencies_to_current_module_and_install_them) to add the github.com/gin-gonic/gin module as a dependency for your module. Use a dot argument to mean "get dependencies for code in the current directory."



```shell
$ go get .
go get: added github.com/gin-gonic/gin v1.7.2
```



From the command line in the directory containing main.go, run the code. Use a dot argument to mean "run code in the current directory."

```shell
$ go run .
```



From a new command line window, use `curl` to make a request to your running web service.

```shell
$ curl http://localhost:8080/albums
```



```shell
$ curl http://localhost:8080/albums \
    --header "Content-Type: application/json" \
    --request "GET"
```



post data

```shell
curl http://localhost:8080/albums    --include    --header "Content-Type: application/json"   --request "POST"   --data "{\"id\": \"4\",\"title\": \"The Modern Sound of Betty Carter\",\"artist\": \"Betty Carter\",\"price\": 49.99}"
```





reference:

[Tutorial: Developing a RESTful API with Go and Gin - The Go Programming Language (golang.org)](https://golang.org/doc/tutorial/web-service-gin)