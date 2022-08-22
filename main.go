package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	//"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"text/template"
	"time"
)

func main() {
	var pathFlag = flag.String("path", "/var/www/schule.3nt3.de/schule", "Path to scan - '.../.../schule'")
	flag.Parse()

	type entry struct {
		DateString    string
		Title         string
		IsFile        bool
		LatexPath     string
		PDFPath       string
		DirectoryPath string
	}

	var latexFiles []string
	var directories []string

	dateRe := regexp.MustCompile("[0-9]{4}-[0-9]{2}-[0-9]{2}")

	err := filepath.Walk(*pathFlag,
		func(path string, info os.FileInfo, err error) error {
			log.Printf("%+v\n", info)
			if info.IsDir() && dateRe.MatchString(info.Name()) {
				directories = append(directories, path)
			}
			return nil
		})

	filteredDirectories := []string{}
	for _, directory := range directories {
		var foundLaTeX bool
		err = filepath.Walk(directory,
			func(path string, info os.FileInfo, err error) error {
				if err != nil {
					log.Printf("err: %v\n", err)
					return err
				}

				if strings.HasSuffix(path, ".tex") {
					latexFiles = append(latexFiles, path)
					foundLaTeX = true
				}

				return nil
			})

		if !foundLaTeX {
			filteredDirectories = append(filteredDirectories, directory)
		}
	}

	re, err := regexp.Compile("\\\\title{(.*?)}")
	if err != nil {
		log.Printf("error in regex: %v\n", err)
		return
	}

	entries := make(map[string][]entry, 0)
	for _, directory := range filteredDirectories {
		// FIXME: change to other path
		path := strings.TrimPrefix(directory, *pathFlag)

		splitPath := strings.Split(path, "/")
		if strings.HasSuffix(path, "__latexindent_temp.tex") {
			continue
		}

		log.Printf("%v\n", splitPath[3])

		// 0 - Q1
		// 1 - Q1/physik
		// 2 - Q1/physik/2021-12-07
		dateString := splitPath[3]
		subjectString := splitPath[2]
		if subjectString == "misc" {
			continue
		}

		// respect hidden files
		if strings.HasPrefix(subjectString, ".") || strings.HasPrefix(dateString, ".") {
			continue
		}

		if entries[subjectString] == nil {
			entries[subjectString] = []entry{}
		}

		entries[subjectString] = append(entries[subjectString], entry{DateString: dateString, Title: "", LatexPath: "", PDFPath: "", DirectoryPath: "/schule" + path, IsFile: false})
	}

	for _, path := range latexFiles {
		b, err := ioutil.ReadFile(path)
		if err != nil {
			log.Printf("error reading file: %v\n", err)
			continue
		}

		matches := re.FindAllString(string(b), -1)
		if len(matches) > 0 {
			title := regexp.MustCompile("(\\\\title{|})").ReplaceAllString(matches[0], "")

			// FIXME: change path
			path := strings.TrimPrefix(path, *pathFlag)
			if strings.HasSuffix(path, "__latexindent_temp.tex") {
				continue
			}

			splitPath := strings.Split(path, "/")
			l := len(splitPath)
			dateString := splitPath[3]
			subjectString := splitPath[2]
			if subjectString == "misc" {
				continue
			}

			if entries[subjectString] == nil {
				entries[subjectString] = []entry{}
			}

			filenameSplit := strings.Split(splitPath[l-1], ".")

			filenameSplit[len(filenameSplit)-1] = "pdf"
			pdfFile := strings.Join(filenameSplit, ".")

			splitPath[l-1] = pdfFile
			pdfPath := strings.Join(splitPath, "/")

			entries[subjectString] = append(entries[subjectString], entry{DateString: dateString, Title: title, LatexPath: path, PDFPath: pdfPath, IsFile: true})
		}
	}

	//var sortedEntries = make(map[string][]entry, 0)

	// sort by date
	for subject := range entries {
		sort.SliceStable(entries[subject], func(i, j int) bool {
			ei := entries[subject][i]
			ej := entries[subject][j]

			return ei.DateString < ej.DateString
		})
	}

	b, err := ioutil.ReadFile("template.html")
	if err != nil {
		log.Printf("error reading template.html: %v\n", err)
		return
	}

	t, err := template.New("lol").Parse(string(b))
	if err != nil {
		log.Printf("error templating: %v\n", err)
		return
	}

	var subjects []string
	for k := range entries {
		subjects = append(subjects, k)
	}

	sort.Strings(subjects)

	//version, err := exec.Command("sh", "-c", "git", "rev-parse", "--short", "HEAD").Output()
	version := ""
	if err != nil {
		log.Fatal(err)
	}

	err = t.Execute(os.Stdout, map[string]interface{}{"subjects": subjects, "entries": entries, "lastUpdated": time.Now(), "version": version})
	if err != nil {
		log.Printf("error templating some more: %v\n", err)
		return
	}
}
