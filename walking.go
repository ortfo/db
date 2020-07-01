package main

import (
	"fmt"
	"io/ioutil"
	"path"
)

type ProjectTreeElement struct {
	ID             string
	DescriptionRaw string
	MediaFilepaths []string
}

func BuildProjectsTree(databaseDirectory string) ([]ProjectTreeElement, error) {
	var tree []ProjectTreeElement
	files, err := ioutil.ReadDir(databaseDirectory)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		fmt.Println("Scanning " + file.Name())
		// If it's not a directory, it's not a project folder
		// so it has nothing to do with this
		if !file.IsDir() {
			continue
		}
		// If the directory has no description.md, it's not a project folder
		descriptionFilepath := path.Join(databaseDirectory, file.Name(), "description.md")
		if !FileExists(descriptionFilepath) {
			continue
		}
		// Read the description.md file
		descriptionRaw := ReadFile(descriptionFilepath)
		
		// Build the list of media filepaths
		mediaFiles, err := ioutil.ReadDir(path.Join(databaseDirectory, file.Name()))
		var mediaFilepaths []string
		if err != nil {
			return nil, err
		}
		for _, mediaFile := range mediaFiles {
			if mediaFile.Name() == "description.md" {
				continue
			}
			mediaFilepaths = append(mediaFilepaths, mediaFile.Name())
		}
		// Append the new project
		tree = append(tree, ProjectTreeElement{ID: file.Name(), DescriptionRaw: descriptionRaw, MediaFilepaths: mediaFilepaths})
	}
	return tree, nil
}
