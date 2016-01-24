/*
Copyright: master_hilli

    This program is free software: you can redistribute it and/or modify
    it under the terms of the GNU General Public License as published by
    the Free Software Foundation version 3 of the License, or
    any later version.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU General Public License for more details.

    You should have received a copy of the GNU General Public License
    along with this program.  If not, see <http://www.gnu.org/licenses/>.

*/
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
const POSSIBLE_ARG_COUNT int = 2
const POS_MAIN_PACKAGE_NAME int = 1
var packageNameForGoTool string = `github.com`

func main() {
    retrieveMainPackageName()
    currentPath,_ := filepath.Abs(".")
    contentOfCoverFiles := "mode: set\n" + createCoverFileForDirectoryRecursive(currentPath)

    switchDirectoryToPath(currentPath)
    writeContentToCodeCoverageFile(contentOfCoverFiles)
    //fmt.Printf("\nCONTENT: \n%s\n", contentOfCoverFiles)
    createCodeCoverageHtmlPage()
}

func retrieveMainPackageName() {
    if (len(os.Args) != POSSIBLE_ARG_COUNT) {
        fmt.Println("Information: no main package name provided. Fallback is 'github.com'")
    } else {
        packageNameForGoTool = os.Args[POS_MAIN_PACKAGE_NAME]
    }
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

    // TODO: create own method for copying -- so I can create an OS independent function
    commandToAllFilesFromCurrentFolder := exec.Command("xcopy.exe", "/Y", "*.go", packageFolderInGoPath)
    output, err := commandToAllFilesFromCurrentFolder.Output()
    if (err != nil) {
        panic(err)
    }
    fmt.Println(string(output))
}

func createCodeCoverageHtmlPage() {
    commandToCreateHTMLFileForCodeCoverage := exec.Command(programName, "tool", "cover", "-html="+codeCoverOutputFileName, "-o", codeCoverOutputFileName+".html")
    _, err := commandToCreateHTMLFileForCodeCoverage.Output()
    if (err != nil) {
        fmt.Printf("ERROR in HTML file generation: %s\n", err.Error())
    } else {
        fmt.Println("Html file was created successfully")
    }
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
    //TODO: need to somehow remove the blanks from the path files, or perhabs that is even not allowed?
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
