## Repository information

This Repository contains all plugins for the [Iris web framework version 5/fasthttp](https://github.com/kataras/iris/tree/5.0.0).

You can contribute also, just make a pull request, try to keep conversion, configuration file: './myplugin/config.go' & plugin: './myplugin/myplugin.go'.


## How can I install a plugin?

```sh
$ go get -u github.com/iris-contrib/plugin.v5/$FOLDERNAME
```

## How can I register a plugin?

```go
iris.Plugins.Add(thePlugin)
// or per-iris instance
app := iris.New()
app.Plugins.Add(thePlugin)
```
