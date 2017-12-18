package main

import (
	"github.com/ghodss/yaml"
	"io/ioutil"
	"log"
	"os"
	"text/template"
	"strings"
	"path/filepath"
	"fmt"
)

type Deployment struct {
	ApiVersion string
	Kind       string
	Name       string
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
		Replicas     int32
		Scalable     bool
		Clustering   bool
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
		LivenessProbe struct {
			HttpGet struct {
				Path                string
				Port                int32
				InitialDelaySeconds int32
				PeriodSeconds       int32
			}
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
	err = template.Execute(outputFile, data)
	if err != nil {
		log.Print("Error executing template:", err)
		return
	}
	outputFile.Close()
}

func main() {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)
	fmt.Println(exPath)

	deployment := getDeployment(exPath + "/depoyment.yaml")
	componentNamesMap := map[string]bool {}

	for _, component := range deployment.Components {
		if _, ok := componentNamesMap[component.CodeName]; ok {
			continue
		}

		componentNamesMap[component.CodeName] = true
		templatePath := exPath+"/templates/docker/Dockerfile.template"
		outputFilePath := exPath+"/output/docker/" + component.CodeName + "/Dockerfile"
		applyTemplate(templatePath, outputFilePath, component)
	}
}
