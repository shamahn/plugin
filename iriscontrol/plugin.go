package iriscontrol

import (
	"os"

	"gopkg.in/kataras/go-fs.v0"
	"gopkg.in/kataras/iris.v4"
	"gopkg.in/kataras/iris.v4/utils"
)

// Name "Iris Control"
const Name = "Iris Control"

var (
	assetsURL       = "https://github.com/iris-contrib/iris-control-assets/archive/4.0.0.zip"
	assetsUnzipname = "iris-control-assets"
	assetsPath      = ""
	workingDir      = ""
)

// init just sets the assetsPath & current workingDir
func init() {
	workingDir, _ = os.Getwd()
	assetsPath = utils.AssetsDirectory + fs.PathSeparator + "iris-control-assets" + fs.PathSeparator
}

func installAssets() {

	if !fs.DirectoryExists(assetsPath) {
		errMsg := "\nProblem while downloading the assets from the internet for the first time. Trace: %s"

		installedDir, err := fs.Install(assetsURL, assetsPath, true)
		if err != nil {
			panic(errMsg)
		}

		err = fs.CopyDir(installedDir, assetsPath)
		if err != nil {
			panic(err)
		}

		// try to remove the unzipped folder
		fs.RemoveFile(installedDir[0 : len(installedDir)-1])
	}
}

// New creates & returns a new iris control plugin
// receives two parameters
// first is the authenticated users which should be able to access the control panel
// second is the PORT which the iris control panel should be listened & served to
func New(port int, users map[string]string) IrisControl {
	return &iriscontrol{port: port, users: users}
}

// PreListen registers the iriscontrol plugin
func (i *iriscontrol) PreListen(s *iris.Framework) {
	installAssets()
	i.listen(s)
}

// GetName returns the name of the plugin
func (i *iriscontrol) GetName() string {
	return Name
}

// GetDescription returns the description of the plugin
func (i *iriscontrol) GetDescription() string {
	return Name + " is just a web interface which gives you control of your Iris.\n"
}

// PreClose any clean-up
// temporary is empty because all resources are cleaned graceful by the iris' station
func (i *iriscontrol) PreClose(s *iris.Framework) {}
