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
    "io"
    "os"
    "os/exec"
    "path/filepath"
    "bufio"
    "strings"
    "io/ioutil"
    "regexp"
)

// printed information to user:
const usageInformationToPrint string = `Information: no main package name provided. Fallback is 'github.com'.
If you want to use it for your specific package (meaning first package name under %GOPATH%/src/<yourpackage>)
usage:
goCodeCoverHtmlGenerator <yourpackage>
`

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

// regexp constants
const REGEXP_GO_FILES string = ".*\\.go"
const REGEXP_GO_TEST_FILES string = ".*test\\.go"

func main() {
    retrieveMainPackageName()
    currentPath,_ := filepath.Abs(".")
    contentOfCoverFiles := "mode: set\n" + createCoverFileForDirectoryRecursive(currentPath)

    switchDirectoryToPath(currentPath)
    writeContentToCodeCoverageFile(contentOfCoverFiles)
    createCodeCoverageHtmlPage()
}

func retrieveMainPackageName() {
    if (len(os.Args) != POSSIBLE_ARG_COUNT) {
        fmt.Println(usageInformationToPrint)
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
        copyNewGoFilesToGoRootSrcWhenInSeparateLocation()
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
    paths := getGoFilePathsFromDirectory(path, REGEXP_GO_TEST_FILES)
    return len(paths) > 0
}

func getGoFilePathsFromDirectory(path string, regexForFile string) []string{
    fileInfo, err := ioutil.ReadDir(path)
    if err != nil {
        panic(err)
    }
    var names []string = make([]string, 0)
    path = addSeparator(path)
    for i := range fileInfo {
        if !fileInfo[i].IsDir() {
            matchesGoFile, _ := regexp.MatchString(regexForFile, fileInfo[i].Name())

            if matchesGoFile {
                names = append(names, fileInfo[i].Name())
            }
        }
    }
    return names
}

func createCoverageFile() string{
    executeTestWithCoverageInCurrentFolder()
    codeCoverageFile := openCodeCoverageOutputFile()
    defer codeCoverageFile.Close()
    relativeCodeCoverFileContent := makePathsRelativeForContentIn(codeCoverageFile)
    codeCoverageFile.Close()
    writeContentToCodeCoverageFile("mode: set\n" + relativeCodeCoverFileContent)
    return relativeCodeCoverFileContent
}

func copyNewGoFilesToGoRootSrcWhenInSeparateLocation() {
    // TODO create method get the package go import path
    var relativePathToCurrentPackage string
    currentPath,_ := filepath.Abs(".")
    indexForPackageStart := strings.Index(currentPath, packageNameForGoTool)
    if (indexForPackageStart > 0) {
        relativePathToCurrentPackage = currentPath[indexForPackageStart:len(currentPath)]
    } else {
        fmt.Printf("ERROR: Could not retrieve index for provided package in current path.\n\tpath:\t%s\n\tpkg:\t%s\n",
                   currentPath, packageNameForGoTool)
        return
    }

    goPathSrc := createPathToGoPathSrc()
    if(!isCurrentExecutionPathAlreadyInGoRootSrc(currentPath, goPathSrc)) {
        // TODO own method to create the path to package under libraries
        packageFolderInEnvVariableGoPath := goPathSrc + relativePathToCurrentPackage
        fmt.Printf("current: %s\ngoPathSrc: %s\n", currentPath, packageFolderInEnvVariableGoPath)
        err := os.MkdirAll(packageFolderInEnvVariableGoPath, 0777)
        if err != nil { panic(err)}


        copyFilesToGOROOTPath(packageFolderInEnvVariableGoPath)
    }
}

func createPathToGoPathSrc() string {
    goPathSrc := getEnvironmentVariable(gopathKey)
    goPathSrc = addSeparator(goPathSrc) + "src"
    return addSeparator(goPathSrc)
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
        fmt.Printf("Output produced: %s \n ERROR: %s\n", string(output), err.Error())
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

func copyFilesToGOROOTPath(packageFolderInEnvVariableGoPath string) {
    filenames := getGoFilePathsFromDirectory(".", REGEXP_GO_FILES)
    for i := range filenames {
        cpFrom := addSeparator(".") + filenames[i]
        cpTo := addSeparator(packageFolderInEnvVariableGoPath) + filenames[i]
        err := copyFile(cpFrom, cpTo)
        if (err != nil) {
            panic(err)
        }

    }
}

func isCurrentExecutionPathAlreadyInGoRootSrc(currentPath, goPathSrc string) bool {
    return strings.Index(currentPath, goPathSrc) == 0
}


// Copy file content:
// CopyFile copies a file from src to dst. It always overrides the content
// copied and changed that piece of code from: http://stackoverflow.com/questions/21060945/simple-way-to-copy-a-file-in-golang
// by markc
func copyFile(src, dst string) (err error) {
    sfi, err := os.Stat(src)
    if err != nil {
        panic (err)
        return
    }
    if !sfi.Mode().IsRegular() {
        // cannot copy non-regular files (e.g., directories,
        // symlinks, devices, etc.)
        return fmt.Errorf("CopyFile: non-regular source file %s (%q)", sfi.Name(), sfi.Mode().String())
    }
    _, err = os.Stat(dst)
    if err != nil {
        if !os.IsNotExist(err) {
            panic(err)
            return
        }
    }
    err = copyFileContents(src, dst)
    return
}

// copyFileContents copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file.
func copyFileContents(src, dst string) (err error) {
    in, err := os.Open(src)
    if err != nil {
        return
    }
    defer in.Close()
    out, err := os.Create(dst)
    if err != nil {
        return
    }
    defer func() {
        cerr := out.Close()
        if err == nil {
            err = cerr
        }
    }()
    if _, err = io.Copy(out, in); err != nil {
        return
    }
    err = out.Sync()
    return
}
