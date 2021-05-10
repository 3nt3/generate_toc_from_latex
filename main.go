package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"text/template"
)

func main() {
	type entry struct {
		DateString string
		Title      string
		LatexPath  string
		PDFPath    string
	}

	var latexFiles []string

	err := filepath.Walk("/var/www/schule.3nt3.de/schule",
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				log.Printf("err: %v\n", err)
				return err
			}

			if strings.HasSuffix(path, ".tex") {
				latexFiles = append(latexFiles, path)
			}

			return nil
		})

	re, err := regexp.Compile("\\\\title{(.*?)}")
	if err != nil {
		log.Printf("error in regex: %v\n", err)
		return
	}

	entries := make(map[string][]entry, 0)
	for _, path := range latexFiles {
		b, err := ioutil.ReadFile(path)
		if err != nil {
			log.Printf("error reading file: %v\n", err)
			continue
		}

		matches := re.FindAllString(string(b), -1)
		if len(matches) > 0 {
			title := regexp.MustCompile("(\\\\title{|})").ReplaceAllString(matches[0], "")

			path := strings.TrimPrefix(path, "/var/www/schule.3nt3.de")
			if strings.HasSuffix(path, "__latexindent_temp.tex") {
				continue
			}

			splitPath := strings.Split(path, "/")
			l := len(splitPath)
			dateString := splitPath[l-2]
			subjectString := splitPath[l-3]
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

			entries[subjectString] = append(entries[subjectString], entry{DateString: dateString, Title: title, LatexPath: path, PDFPath: pdfPath})
		}
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

	subjects := []string{}
	for k := range entries {
		subjects = append(subjects, k)
	}

	sort.Strings(subjects)

	err = t.Execute(os.Stdout, map[string]interface{}{"subjects": subjects, "entries": entries})
	if err != nil {
		log.Printf("error templating some more: %v\n", err)
		return
	}
}
