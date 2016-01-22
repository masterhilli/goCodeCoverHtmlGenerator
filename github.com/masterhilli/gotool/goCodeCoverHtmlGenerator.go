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
    createCoverageFile()
    copyNewGoFilesToLibrary()
    createCodeCoverageHtmlPage()
}

func createCoverageFile() {
    executeTestWithCoverageInCurrentFolder()
    codeCoverageFile := openCodeCoverageOutputFile()
    relativeCodeCoverFileContent := makePathsRelativeForContentIn(codeCoverageFile)
    codeCoverageFile.Close()
    writeContentToCodeCoverageFile(relativeCodeCoverFileContent)
}

func copyNewGoFilesToLibrary() {
    // TODO create method get the package go import path
    var pathToCopyToGoRootFolder string
    currentPath,_ := filepath.Abs(".")
    indexForPackageStart := strings.Index(currentPath, packageNameForGoTool)
    if (indexForPackageStart > 0) {
        pathToCopyToGoRootFolder = currentPath[indexForPackageStart:len(currentPath)]
    }

    // TODO own method to create the path to package under libraries
    goPath := getEnvironmentVariable(gopathKey)
    packageFolderInGoPath := goPath + string(filepath.Separator) + "src" +string(filepath.Separator) + pathToCopyToGoRootFolder
    err := os.MkdirAll(packageFolderInGoPath, 0777)
    if err != nil { panic(err)}

    // todo create own method for copying -- so I can create an OS independent function
    commandToAllFilesFromCurrentFolder := exec.Command("xcopy.exe", "/Y", "*.go", packageFolderInGoPath)
    output, err := commandToAllFilesFromCurrentFolder.Output()
    if (err != nil) {
        panic(err)
    }
    fmt.Println(string(output))
}

func createCodeCoverageHtmlPage() {
    commandToCreateHTMLFileForCodeCoverage := exec.Command(programName, "tool", "cover", "-html="+codeCoverOutputFileName, "-o", codeCoverOutputFileName+".html")
    output, err := commandToCreateHTMLFileForCodeCoverage.Output()
    if (err != nil) {
        panic(err)
    }
    fmt.Println(output)
    // TODO: execute command to create HTML file and start default browser with HTML page
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

func openCodeCoverageOutputFile() *os.File {
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

