## Repository information

This Repository contains all plugins for the [Iris web framework v4](https://github.com/kataras/iris/tree/4.0.0).

You can contribute also, just make a pull request, try to keep conversion, configuration file: './myplugin/config.go' & plugin: './myplugin/myplugin.go'.


## How can I install a plugin?

```sh
$ go get -u gopkg.in/iris-contrib/plugin.v4/$FOLDERNAME
```

## How can I register a plugin?

```go
iris.Plugins.Add(theplugin)
```
