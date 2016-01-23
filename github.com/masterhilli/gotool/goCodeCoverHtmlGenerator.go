package main

import (
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "bufio"
    "strings"
    "io/ioutil"
    "regexp"
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
    currentPath,_ := filepath.Abs(".")
    contentOfCoverFiles := "mode: set\n" + createCoverFileForDirectoryRecursive(currentPath)

    switchDirectoryToPath(currentPath)
    writeContentToCodeCoverageFile(contentOfCoverFiles)
    //fmt.Printf("\nCONTENT: \n%s\n", contentOfCoverFiles)
    createCodeCoverageHtmlPage()
}
func createCoverFileForDirectoryRecursive(path string) string{
    var contentOfCoverFile string = ""
    if (!filepath.IsAbs(path)) {
        fmt.Printf("Warning: path is not an absolute path (%s)", path)
    }
    childDirectories := getDirectoriesOfPath(path)
    for i := range childDirectories {
        contentOfCoverFile = contentOfCoverFile + createCoverFileForSubDirectoryRecursive(path, childDirectories[i])
    }
    if directoryHasGoTestFiles(path) {
        // TODO: change current execution directory to new directory!
        switchDirectoryToPath(path)
        contentOfCoverFile = contentOfCoverFile + createCoverageFile()
        copyNewGoFilesToLibrary()
    }
    return contentOfCoverFile
}

func switchDirectoryToPath(path string) {
    err := os.Chdir(path)
    if (err != nil) {
        panic(err)
    }
}

func createCoverFileForSubDirectoryRecursive(path, childDirectory string) string {
    path = addSeparator(path)
    path = addSeparator(path + childDirectory)
    return createCoverFileForDirectoryRecursive(path)
}

func addSeparator(path string) string {
    if (len(path) > 0 && strings.LastIndex(path, string(filepath.Separator)) != len(path)-1) {
        path = path + string(filepath.Separator)
    }
    return path
}

func getDirectoriesOfPath(path string) []string{
    var directories []string = make([]string, 0)
    fileInfo, err := ioutil.ReadDir(path)
    if err != nil {
        panic(err)
    }

    for i := range fileInfo {
        if fileInfo[i].IsDir() {
            directories = append(directories, fileInfo[i].Name())
        }
    }
    return directories
}

func directoryHasGoTestFiles(path string) bool {
    fileInfo, err := ioutil.ReadDir(path)
    if err != nil {
        panic(err)
    }

    for i := range fileInfo {
        if !fileInfo[i].IsDir() {
            matchGoTest, _ := regexp.MatchString(".*test\\.go", fileInfo[i].Name())
            if matchGoTest {
                return true
            }
        }
    }
    return false
}

func createCoverageFile() string{
    executeTestWithCoverageInCurrentFolder()
    codeCoverageFile := openCodeCoverageOutputFile()
    relativeCodeCoverFileContent := makePathsRelativeForContentIn(codeCoverageFile)
    codeCoverageFile.Close()
    writeContentToCodeCoverageFile("mode: set\n" + relativeCodeCoverFileContent)
    return relativeCodeCoverFileContent
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
    fmt.Println(string(output))
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
    line, isPrefix, err := codeCoverageReader.ReadLine() // I need to skip the first line! module set will be added later
    line, isPrefix, err = codeCoverageReader.ReadLine()
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

