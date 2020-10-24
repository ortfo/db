package main

import (
	"io/ioutil"
	"path"
)


// ProjectTreeElement represents a project
type ProjectTreeElement struct {
	ID             string
	DescriptionRaw string
	MediaFilepaths []string
}

// GetProjectPath returns the project's folder path with regard to databaseDirectory
func (p *ProjectTreeElement) GetProjectPath(databaseDirectory string) string {
	return path.Join(".", databaseDirectory, p.ID)
}

// MediaAbsoluteFilepaths is like MediaFilepaths but returns absolute paths with regard to databaseDirectory
func (p *ProjectTreeElement) MediaAbsoluteFilepaths(databaseDirectory string) []string {
	absoluted := make([]string, len(p.MediaFilepaths))
	for _, item := range p.MediaFilepaths {
		absoluted = append(absoluted, path.Join(p.GetProjectPath(databaseDirectory), item))
	}
	return absoluted
}

// BuildProjectsTree scans databaseDirectory to return a slice of ProjectTreeElement's, gathering media files and other various information
func BuildProjectsTree(databaseDirectory string) ([]ProjectTreeElement, error) {
	var tree []ProjectTreeElement
	files, err := ioutil.ReadDir(databaseDirectory)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
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
