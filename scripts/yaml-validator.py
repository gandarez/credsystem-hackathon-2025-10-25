import yaml
from dataclasses import dataclass

with open('example.yaml', 'r') as file:
    config = yaml.safe_load(file)

print(config)


@dataclass
class DockerComposeYaml2:
    backend_api: 
    


@dataclass
class DockerComposeYaml:
    services: DockerComposeYaml2
