# Gerador

Gerador is an infrastructure resource generator. It makes use of a generic deployment specification for defining the
deployment architecture of any software system and generating resources required for deploying the given software system
on Docker, Docker Compose, Kubernetes, OpenShift, Pivotal Cloud Foundry, DC/OS, AWS, Azure and Google Cloud.

## Deployment Specification

````yaml
specVersion: 0.1
kind: Deployment
name: Name of the deployment
version: Version of the deployment
components:  # List of components
-
  name: Name of the component
  codeName: Code name of the component
  version: Version of the component
  cpu: Number of CPUs required
  memory: Amount of memory required 
  disk: Amount of disk space required
  distribution: Distribution file name
  entrypoint: Startup script
  replicas: Number of replicas
  scalable: Scalable or not
  clustering: Clustering needed or not
  ports:
  -
    name: Port name
    protocol: Protocol of the port
    port: Port number exposed
    external: Port need to be exposed externally or not
    sessionAffinity: Session affinity required or not
  databases:
  -
    name: Database name
    createScript: Path of the database creation script
  dependencies:
  -
    component: Dependent component code name
    ports:
    - Name of the dependent component port used
  livenessProbe:
    httpGet:
      path: Context path of the HTTP endpoint
      port: Port of the HTTP endpoint
      initialDelaySeconds: Initial delay in seconds
      periodSeconds: Period in seconds
````