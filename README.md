## Repository information

This Repository contains all plugins for the [Iris web framework](https://github.com/kataras/iris).

> For Iris v5/fasthttp plugins please navigate [here](https://github.com/iris-contrib/plugin/tree/5.0.0).

You can contribute also, just make a pull request, try to keep conversion, configuration file: './myplugin/config.go' & plugin: './myplugin/myplugin.go'.


## How can I install a plugin?

```sh
$ go get -u github.com/iris-contrib/plugin/$FOLDERNAME
```

## How can I register a plugin?

```go
iris.Plugins.Add(thePlugin)
// per iris instance
app := iris.New()
app.Plugins.Add(thePlugin)
```
