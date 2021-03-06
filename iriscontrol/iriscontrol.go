package iriscontrol

import (
	"strconv"
	"time"

	"github.com/iris-contrib/middleware/basicauth"
	"github.com/kataras/go-template/html"
	"github.com/kataras/iris"
)

type (
	// IrisControl is the interface which the iriscontrol should implements
	// it's empty for now because no need any public API
	IrisControl interface{}
	iriscontrol struct {
		port  int
		users map[string]string

		// child is the plugin's standalone station
		child *iris.Framework
		// the station which this plugins is registed to
		parent       *iris.Framework
		parentLastOp time.Time

		// websocket
		clients clients
	}

	clients []iris.WebsocketConnection

	pluginInfo struct {
		Name        string
		Description string
	}

	logInfo struct {
		Date      string
		Status    int
		Latency   time.Duration
		IP        string
		Method    string
		Subdomain string
		Path      string
	}
)

func (c clients) indexOf(connectionID string) int {
	for i := range c {
		if c[i].ID() == connectionID {
			return i
		}
	}
	return -1
}

var _ IrisControl = &iriscontrol{}

func (i *iriscontrol) listen(f *iris.Framework) {
	// set the path logger to the parent which will send the log via websocket to the browser
	f.UseGlobalFunc(func(ctx *iris.Context) {
		status := ctx.ResponseWriter.StatusCode()
		path := ctx.Path()
		method := ctx.Method()
		subdomain := ctx.Subdomain()
		ip := ctx.RemoteAddr()
		startTime := time.Now()

		ctx.Next()
		//no time.Since in order to format it well after
		endTime := time.Now()
		date := endTime.Format("01/02 - 15:04:05")
		latency := endTime.Sub(startTime)
		info := logInfo{
			Date:      date,
			Status:    status,
			Latency:   latency,
			IP:        ip,
			Method:    method,
			Subdomain: subdomain,
			Path:      path,
		}
		i.Emit("log", info) //send this text to the browser,
	})

	i.parent = f
	i.parentLastOp = time.Now()

	i.initializeChild()
}

func (i *iriscontrol) initializeChild() {
	i.child = iris.New()
	i.child.Config.DisableBanner = true
	i.child.UseTemplate(html.New()).Directory(assetsPath+"templates", ".html")
	i.child.Config.Websocket.Endpoint = "/ws"

	// set the assets
	i.child.StaticWeb("/public", assetsPath+"static")

	// set the authentication middleware to all except websocket
	auth := basicauth.New(basicauth.Config{
		Users:      i.users,
		ContextKey: "user",
		Realm:      basicauth.DefaultBasicAuthRealm,
		Expires:    time.Duration(1) * time.Hour,
	})

	i.child.UseFunc(func(ctx *iris.Context) {
		///TODO: Remove this and make client-side basic auth when websocket connection. (user@password/host.. on chronium)
		// FOR GOOGLE CHROME/CHRONIUM
		// https://bugs.chromium.org/p/chromium/issues/detail?id=123862
		// CROSS DOMAIN IS DISABLED so I think this is ok solution for now...
		if ctx.Path() == i.child.Config.Websocket.Endpoint {
			ctx.Next()
			return
		}
		auth.Serve(ctx)
	})

	i.child.Websocket.OnConnection(func(c iris.WebsocketConnection) {
		// add the client to the list
		i.clients = append(i.clients, c)
		c.OnDisconnect(func() {
			// remove the client from the list
			if idx := i.clients.indexOf(c.ID()); idx != -1 {
				i.clients[idx] = i.clients[len(i.clients)-1]
				i.clients = i.clients[:len(i.clients)-1]
			}

		})
	})

	i.child.Get("/", func(ctx *iris.Context) {
		ctx.MustRender("index.html", iris.Map{
			"ServerIsRunning":      i.parentIsRunning(),
			"Host":                 i.child.Config.VHost,
			"Routes":               i.parentLookups(),
			"Plugins":              i.infoPlugins(),
			"LastOperationDateStr": i.infoLastOp(),
		})
	})

	i.child.Post("/start_server", func(ctx *iris.Context) {

		if !i.parent.IsRunning() {
			// starts the server with its old configuration
			go func() {
				if err := i.parent.Reserve(); err != nil {
					i.parent.Logger.Println(err.Error())
				}
			}()

			i.parentLastOp = time.Now()
		}
	})

	i.child.Post("/stop_server", func(ctx *iris.Context) {

		if i.parentIsRunning() {
			i.parentLastOp = time.Now()

			go func() {
				if err := i.parent.Close(); err != nil {
					i.parent.Logger.Println(err.Error())
				}
			}()
		}
	})

	go i.child.Listen(iris.ParseHostname(i.parent.Config.VHost) + ":" + strconv.Itoa(i.port))
}

func (i *iriscontrol) parentIsRunning() bool {
	return i.parent != nil && i.parent.IsRunning()
}

func (i *iriscontrol) parentLookups() []iris.Route {
	if i.parent == nil {
		return nil
	}
	return i.parent.Lookups()
}

func (i *iriscontrol) infoPlugins() (info []pluginInfo) {
	plugins := i.parent.Plugins
	for _, p := range plugins.GetAll() {
		name := plugins.GetName(p)
		description := plugins.GetDescription(p)
		if name == "" {
			name = "Unknown plugin name"
		}
		if description == "" {
			description = "description is not available"
		}

		info = append(info, pluginInfo{Name: name, Description: description})
	}
	return
}

func (i *iriscontrol) infoLastOp() string {
	return i.parentLastOp.Format(TimeFormat)
}

func (i *iriscontrol) Emit(event string, msg interface{}) {
	for j := range i.clients {
		i.clients[j].Emit(event, msg)
	}
}
