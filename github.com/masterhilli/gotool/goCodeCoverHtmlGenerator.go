package main

import (
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "bufio"
    "strings"
    "io/ioutil"
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
const packageNameForGoTool string = `github.com`

func main() {
    fmt.Println(getEnvironmentVariable(gopathKey))
    executeTestWithCoverageInCurrentFolder()
    codeCoverageFile := openCodecoverageOutputFile()
    relativeCodeCoverFileContent := makePathsRelativeForContentIn(codeCoverageFile)
    codeCoverageFile.Close()
    writeContentToCodeCoverageFile(relativeCodeCoverFileContent)
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
    file, err := os.OpenFile(absCodeCoverOutputFileName, os.O_RDWR, 0600)
    if err != nil {
        panic(err)
    }
    return file
}

func makePathsRelativeForContentIn(codeCoverageFile *os.File) string{
    var relativeFormatedCodeCoverageFileContent string = ""
    codeCoverageReader := bufio.NewReader(codeCoverageFile)
    line, isPrefix, err := codeCoverageReader.ReadLine()
    for err == nil && !isPrefix {
        lineAsString := string(line)
        indexOfPackageStart := strings.Index(lineAsString, packageNameForGoTool)
        if (indexOfPackageStart >= 0) {
            lineAsString = lineAsString[indexOfPackageStart:len(lineAsString)]
        }
        relativeFormatedCodeCoverageFileContent = relativeFormatedCodeCoverageFileContent + lineAsString+ "\n"
        line, isPrefix, err = codeCoverageReader.ReadLine()
    }
    codeCoverageReader.Reset(codeCoverageReader)
    return relativeFormatedCodeCoverageFileContent
}

func writeContentToCodeCoverageFile(relativeCodeCoverFileContent string) {
    err :=  ioutil.WriteFile(codeCoverOutputFileName, []byte(relativeCodeCoverFileContent), 0777)

    if err != nil {
        panic(err)
    }
}