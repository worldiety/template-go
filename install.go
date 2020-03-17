// +builda ignore

package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/worldiety/tools"
	"go/format"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Config provides self-populating getters
type Config struct {
	pkg        *tools.Package
	isDev      bool
	baseURL    string
	projectDir string
	binaryName *string
	isLibrary  *bool
	mainPath   *string
	reader     *bufio.Reader
}

func (c *Config) RootPackageName() string {
	return c.pkg.Name
}

func (c *Config) BaseURL() string {
	return c.baseURL
}

func (c *Config) ProjectDir() string {
	return c.projectDir
}

func (c *Config) ArtifactName() string {
	lib := c.IsLibrary()
	name := c.BinaryName()
	if lib {
		name = "lib" + strings.ToUpper(name[0:1]) + name[1:]
	}
	return name
}

func (c *Config) BinaryName() string {
	if c.binaryName == nil {
		tmp := readString(c.reader, "enter the name of the executable binary", filepath.Base(c.pkg.ImportPath))
		c.binaryName = &tmp
	}
	return *c.binaryName
}

func (c *Config) IsApp() bool {
	return !c.IsLibrary()
}

func (c *Config) IsLibrary() bool {
	if c.isLibrary == nil {
		tmp := accept(c.reader, "project is a library only")
		c.isLibrary = &tmp
	}
	return *c.isLibrary
}

func (c *Config) ModulePath() string {
	return c.pkg.ImportPath
}

func (c *Config) MainPath() string {
	if c.mainPath == nil {
		lib := c.IsLibrary()
		var tmp = ""
		if lib {
			tmp = c.ModulePath()
		} else {
			if c.pkg.Name == "main" {
				tmp = readString(c.reader, "enter the import path of the main package", c.pkg.ImportPath)
			} else {
				tmp = readString(c.reader, "enter the import path of the main package (e.g. "+c.pkg.ImportPath+"/cmd)", "")
			}

		}
		c.mainPath = &tmp
	}
	return *c.mainPath
}

func (c *Config) Apply(resName string) string {
	data := c.download(resName)
	tmpl, err := template.New("tmpl").Parse(string(data))
	check("apply template "+resName, err)
	tmp := &strings.Builder{}
	err = tmpl.Execute(tmp, c)
	check("execute template from "+resName, err)
	return tmp.String()
}

func (c *Config) ApplyGo(url string) string {
	text := c.Apply(url)
	data, err := format.Source([]byte(text))
	check("gofmt "+url, err)
	return string(data)
}

// config helpers

func (c *Config) download(resName string) []byte {
	if c.isDev {
		data, err := ioutil.ReadFile(filepath.Join(c.projectDir, resName))
		check("load dev resource "+resName, err)
		return data
	}
	res, err := http.Get(c.baseURL + resName)
	check("get url content", err)
	data, err := ioutil.ReadAll(res.Body)
	check("read body", err)
	if res.StatusCode != http.StatusOK {
		fmt.Println("download content from " + c.baseURL + resName)
		fmt.Println(string(data))
		os.Exit(-1)
	}
	return data
}

func check(msg string, err error) {
	if err != nil {
		fmt.Printf("cannot %s: %v\n", msg, err)
		panic(err)
	}
}

func accept(reader *bufio.Reader, msg string) bool {
	fmt.Print(msg + " (y/N): ")
	str, err := reader.ReadString('\n')
	str = strings.TrimSpace(str)
	check("cannot read console", err)
	return str == "y" || str == "Y"
}

func readString(reader *bufio.Reader, msg, defaultText string) string {
	fmt.Print(msg)
	if len(defaultText) > 0 {
		fmt.Print(" [" + defaultText + "]")
	}
	fmt.Print(": ")
	str, err := reader.ReadString('\n')
	check("cannot read console", err)
	str = strings.TrimSpace(str)
	if len(str) == 0 {
		str = defaultText
	}
	return str
}

// generator stuff

type Generator struct {
	config *Config
}

func (g *Generator) Write(fname string, data []byte) {
	err := ioutil.WriteFile(filepath.Join(g.config.ProjectDir(), fname), data, os.ModePerm)
	check("write file "+fname, err)
}

func (g *Generator) createMakeFile() {
	text := g.config.Apply("Makefile.tmpl")
	g.Write("Makefile", []byte(text))
}

func (g *Generator) createBuildGoFile() {
	text := g.config.Apply("build.go.tmpl")
	g.Write("build.go", []byte(text))
}

func main() {
	isDev := flag.Bool("dev", false, "set to true for developing this tool and using local resources")
	flag.Parse()

	fmt.Println("welcome to the worldiety guided project setup helper")
	dir, err := os.Getwd()
	check("working directory", err)

	pkg, err := tools.GoList(dir, false)
	if err != nil {
		fmt.Printf("failed to 'go mod list': %v\n", err)
		fmt.Println("this current working directory is not a go module")
		fmt.Println("use 'go mod init my/super/module' first")
		os.Exit(-1)
	}

	reader := bufio.NewReader(os.Stdin)
	g := &Generator{

		config: &Config{
			pkg:        pkg,
			projectDir: dir,
			baseURL:    "https://raw.githubusercontent.com/worldiety/template-go/master/",
			isDev:      *isDev,
			reader:     reader,
		},
	}

	fmt.Printf("your working directory is '%s'\n", dir)
	if *isDev {
		fmt.Println("DEV mode on")
	}

	if accept(reader, "create makefile?") {
		g.createMakeFile()
	}
	g.createBuildGoFile()
}
