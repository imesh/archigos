package main

import (
	"github.com/ghodss/yaml"
	"io/ioutil"
	"log"
	"os"
	"text/template"
	"strings"
	"path/filepath"
	"reflect"
)

type Deployment struct {
	ApiVersion string
	Kind       string
	Name       string
	CodeName   string
	Version    string
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
		Environment  [] string
		Volumes      []string
		Ports [] struct {
			Name            string
			Protocol        string
			Port            int32
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
	log.Println("Reading deployment:", filePath)
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

var fns = template.FuncMap{
	"last": func(x int, a interface{}) bool {
		return x == reflect.ValueOf(a).Len() - 1
	},
}

func applyTemplate(templateFilePath string, outputFilePath string, data interface{}) {
	log.Println("Applying template", templateFilePath)
	template, err := template.ParseFiles(templateFilePath)
	if err != nil {
		log.Print(err)
		return
	}

	lastIndex := strings.LastIndex(outputFilePath, string(os.PathSeparator))
	outputFolderPath := outputFilePath[0: lastIndex]
	os.MkdirAll(outputFolderPath, os.ModePerm);
	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		log.Println("Error creating file:", outputFilePath, err)
		return
	}

	log.Println("Creating file:", outputFilePath)
	template = template.Funcs(fns)
	err = template.Execute(outputFile, data)
	if err != nil {
		log.Print("Error executing template:", err)
		return
	}
	outputFile.Close()
}

func main() {
	exPath := os.Getenv("ARCHIGOS_HOME");
	if (len(exPath) <= 0) {
		ex, err := os.Executable()
		if err != nil {
			panic(err)
		}
		exPath = filepath.Dir(ex)
	}

	deployment := getDeployment(exPath + "/examples/wso2is/depoyment.yaml")
	componentNamesMap := map[string]bool{}

	for _, component := range deployment.Components {
		if _, ok := componentNamesMap[component.CodeName]; ok {
			continue
		}

		componentNamesMap[component.CodeName] = true
		templatePath := exPath + "/templates/docker/Dockerfile.tmpl"
		outputFilePath := exPath + "/output/docker/" + component.CodeName + "/Dockerfile"
		applyTemplate(templatePath, outputFilePath, component)
	}

	templatePath := exPath + "/templates/docker-compose/docker-compose.tmpl"
	outputFilePath := exPath + "/output/docker-compose/" + deployment.CodeName + "/docker-compose.yml"
	applyTemplate(templatePath, outputFilePath, deployment)
}
