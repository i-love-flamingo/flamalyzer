# Flamalyzer - Static code analysis for Flamingo projects

Flamalyzer is a vet-like tool which detects common errors in Flamingo projects.

## Installation

```shell
go install flamingo.me/flamalyzer@latest
```

## Usage

```shell
Flamalyzer [FLAGS] [PATH]
```

### Flags
There are different Flags which can be passed, if no Flags given the Flamalyzer will run in default mode.

```shell
--configFolder=[PATH]
```
To define the path to your config-files, the default is `./.flamalyzer`

```shell
--configSuffix=[SUFFIX]
```

To define a string which must occur in the config-files name which should be loaded 
(useful if you have different files for different use-cases). This only works in combination with ```--configFolder```

```shell
--debugFlamalyzer
```

To enable more information about what's going on in the program. Like info if the given configFolder-Path isn't configured properly.

### Run Flamalyzer within vet

```shell
go vet -vettool=$(which flamalyzer) [PATH]
``` 
 
### Filewatcher
Flamalyzer can be used as external filewatcher to enable error-highlighting in the IDE.
Flamalyzer's output matches exactly vet's output, so the filewatcher configuration from vet can be used. 


## Available Analyses

### Dingo: Pointer receiver check

This analysis checks that an inject-function has a pointer-receiver

### Dingo: correct interface binding check

This analysis checks that an instance implements the interface it is bound to.

### Dingo: proper inject tags check

This analysis checks if the inject tags are used properly.

"inject tags" should be used for annotated injections only (e.g. config), otherwise inject method should be used.

This means:

- No empty inject tags
- Inject tags can be defined in the Inject-Function or must be referenced if defined outside
- They must be declared in the same package as the Inject-Function

#### Architecture: dependency conventions check

This analysis checks all import statements below the entry path that the provided Group-Conventions are respected.

Configuration example:

```yaml
groups:
  entrypaths: [ "src/architecture" ]
  infrastructure:
    allowedDependencies: [ "infrastructure", "interfaces", "application", "domain" ]
  interfaces:
    allowedDependencies: [ "interfaces", "application", "domain" ]
  application:
    allowedDependencies: [ "application", "domain" ]
  domain:
    allowedDependencies: [ "domain" ]
```
 
## Configuration 

The Configuration is done via **yaml**-files.

The files are expected in `./.flamalyzer`

The directory can be specified by `--configFolder=[PATH]`  

**config.yaml** example:
```yaml
# Config of the Dingo-ReceiverAnalyzer
dingoAnalyzer:
  checkPointerReceiver: false
  strictTagsAndFunctions: false
  correctInterfaceToInstanceBinding: false

# Config of the dependencyConventions-ReceiverAnalyzer
dependencyConventionsAnalyzer:
  entryPaths: []
  dependencyConventions: true
  groups:
    infrastructure:
      allowedDependencies: ["infrastructure", "interfaces", "application", "domain"]
    interfaces:
      allowedDependencies: ["interfaces", "application", "domain"]
    application:
      allowedDependencies: ["application", "domain"]
    domain:
      allowedDependencies: ["domain"]
```

There is the possibility to **filter** the files which should be read in.

Use `--configSuffix=[SUFFIX]` to pass a string which must be part of config-file name.
