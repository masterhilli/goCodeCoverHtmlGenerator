# goCodeCoverHtmlGenerator

An executable that tests a package code and the test coverage and if successfull automatically creates the HTML source file!

### Welcome to the Project.

This is a project that creates an executable that can be started in the top folder of your GO-Project and creates an HTML file directly in the started folder, where you can see all your Code Coverage for all GO-Files.
That means you do not have to go to each package to create and view the code coverage, but you can see it in a single html file!

### Features

* Creates the code coverage file for each folder where a *test.go is located
* Copies your packages to your GOPATH (needed for Html Generation by go tool)
* Creates a single HTML file in the executed path with the Code Coverage of all your go files.

### Open Issues

* Copy of files to GOPATH is currently using xcopy of Windows and so the program is not yet compatible with other OS
* Currently does not allow spaces in directories or files

### Authors and Contributors

Author: @masterhilli Year: 2016

### License

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
