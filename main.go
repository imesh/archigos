package main

import (
	"github.com/ghodss/yaml"
	"io/ioutil"
	log "github.com/golang/glog"
	"os"
	"text/template"
	"strings"
	"path/filepath"
	"flag"
	"fmt"
)

const pathSeparator = string(os.PathSeparator)

type Deployment struct {
	ApiVersion string
	Kind       string
	Name       string
	Version    string
	Labels     [] map[string]interface{}
	Components [] struct {
		Name         string
		CodeName     string
		Version      string
		Cpu          string
		Memory       string
		Disk         string
		Distribution string
		Entrypoint   string
		Image        string
		Replicas     int32
		Scalable     bool
		Clustering   bool
		Environment  []string
		Volumes      []string
		Ports [] struct {
			Name            string
			Protocol        string
			Port            int32
			HostPort        int32
			External        bool
			SessionAffinity bool
		}
		Databases [] struct {
			Name         string
			CreateScript string
		}
		Dependencies [] struct {
			Component string
			Ports     [] string
		}
		Healthcheck struct {
			Command struct {
				Assertion []string
			}
			TcpSocket struct {
				Port int32
			}
			HttpGet struct {
				Path string
				Port int32
			}
			InitialDelaySeconds int32
			PeriodSeconds       int32
			TimeoutSeconds      int32
		}
	}
}

func getDeployment(filePath string) *Deployment {
	log.Infoln("Reading deployment:", filePath)
	yamlFile, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Error reading file: %v %v", filePath, err)
	}
	var c *Deployment = new(Deployment)
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		log.Fatalf("Error parsing yaml file: %v %v", filePath, err)
	}
	return c
}

func applyTemplate(templateFilePath string, outputFilePath string, data interface{}) {
	log.V(2).Infoln("Applying template", templateFilePath)
	template, err := template.ParseFiles(templateFilePath)
	if err != nil {
		log.Error(err)
		return
	}

	lastIndex := strings.LastIndex(outputFilePath, string(os.PathSeparator))
	outputFolderPath := outputFilePath[0: lastIndex]
	os.MkdirAll(outputFolderPath, os.ModePerm);
	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		log.Error("Error creating file:", outputFilePath, err)
		return
	}

	log.Infoln("Creating file:", outputFilePath)
	err = template.Execute(outputFile, data)
	if err != nil {
		log.Error("Error executing template:", err)
		return
	}
	outputFile.Close()
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: example -stderrthreshold=[INFO|WARN|FATAL] -log_dir=[string]\n", )
	flag.PrintDefaults()
	os.Exit(2)
}

func init() {
	// Initialize glog
	flag.Usage = usage
	flag.Parse()
	flag.Lookup("logtostderr").Value.Set("true")

}

func generate(executionPath string, deploymentsFolderPath string, filePath string) {
	deployment := getDeployment(filePath)
	componentNamesMap := map[string]bool{}

	// Generate dockerfiles
	for _, component := range deployment.Components {
		// Read component code name
		codeName := component.CodeName
		if codeName == "" {
			codeName = component.Name
		}

		if _, ok := componentNamesMap[codeName]; ok {
			// Dockerfile already generated for component
			continue
		}
		if component.Image != "" {
			// Docker image specified, do not require to generate dockerfile
			continue
		}

		componentNamesMap[codeName] = true
		templatePath := executionPath + pathSeparator + "templates" + pathSeparator + "docker" + pathSeparator + "Dockerfile.tmpl"
		outputFilePath := executionPath + pathSeparator + "output" + pathSeparator + "docker" + pathSeparator + codeName + pathSeparator + "Dockerfile"
		applyTemplate(templatePath, outputFilePath, component)
	}

	// Generate docker compose template
	templatePath := executionPath + pathSeparator + "templates" + pathSeparator + "docker-compose" + pathSeparator + "docker-compose.tmpl"
	outputFilePath := executionPath + pathSeparator + "output" + pathSeparator + "docker-compose"
	// Append sub folder path
	fileFolderPath := strings.Replace(filePath, filepath.Base(filePath), "", 1)
	subFolderPath := strings.Replace(fileFolderPath, deploymentsFolderPath, "", 1)
	if subFolderPath != "" {
		outputFilePath = outputFilePath + subFolderPath + "docker-compose.yml"
	} else {
		outputFilePath = outputFilePath + pathSeparator + "docker-compose.yml"
	}
	applyTemplate(templatePath, outputFilePath, deployment)
}

func main() {
	executionPath := os.Getenv("ARCHIGOS_HOME");
	if (len(executionPath) <= 0) {
		ex, err := os.Executable()
		if err != nil {
			panic(err)
		}
		executionPath = filepath.Dir(ex)
	}

	deploymentFolderPath := executionPath + "/examples"
	log.Infoln("Execution path: ", executionPath)
	log.Infoln("Deployments path: ", deploymentFolderPath)

	err := filepath.Walk(deploymentFolderPath, func(path string, f os.FileInfo, err error) error {
		if !f.IsDir() {
			generate(executionPath, deploymentFolderPath, path)
		}
		return nil
	})
	if err != nil {
		log.Error(err)
	}
}
