package main

import (
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
)

//environment variables
const gopathKey string = "GOPATH"

// execution constants
const programName string = "go"
const boolTestParam string = "test"
const boolCodeCoverParam string = "-cover"
const boolCoverProfile string = "-coverprofile"
const codeCoverOutputFileName string = "cover.out"

var paramsForTestWithCodeCoverage []string = []string{boolTestParam, boolCodeCoverParam, boolCoverProfile, codeCoverOutputFileName}

//packagename
const packageNameForGoTool string = `\github.com\masterhilli\gotool\`

func main() {
    fmt.Println(getEnvironmentVariable(gopathKey))
    executeTestWithCoverageInCurrentFolder()
    codeCoverageFile := openCodecoverageOutputFile()


    defer codeCoverageFile.Close()

}

func getEnvironmentVariable(envKey string) string {
    return os.Getenv(envKey)
}

func executeTestWithCoverageInCurrentFolder() {
    commandToCreateCodeCoverageFile := exec.Command(programName, paramsForTestWithCodeCoverage...)
    output, err := commandToCreateCodeCoverageFile.Output()
    if (err != nil) {
        panic(err)
    }
    fmt.Println(string(output))
}

func openCodecoverageOutputFile() *os.File {
    absCodeCoverOutputFileName, _ := filepath.Abs(codeCoverOutputFileName)
    file, err := os.Open(absCodeCoverOutputFileName)
    if err != nil {
        panic(err)
    }
    return file
}