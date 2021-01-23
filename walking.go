package main

import (
	"io/ioutil"
	"os"
	"path"
)

// ProjectTreeElement represents a project
type ProjectTreeElement struct {
	ID             string
	DescriptionRaw string
	MediaFilepaths []string
	ScatteredMode  bool // Whether the build was run with --scattered
}

// GetProjectPath returns the project's folder path with regard to databaseDirectory
func (p *ProjectTreeElement) GetProjectPath(databaseDirectory string) string {
	if p.ScatteredMode {
		return path.Join(databaseDirectory, p.ID, ".portfoliodb")
	}
	return path.Join(databaseDirectory, p.ID)
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
	for _, projectFolder := range files {
		// If it's not a directory, it's not a project folder
		// so it has nothing to do with this
		if !projectFolder.IsDir() {
			continue
		}
		projectFolderPath := path.Join(databaseDirectory, projectFolder.Name())
		// Read the description.md file
		// If description is empty, then the project is not portfoliodb-enabled.
		// See ReadDescriptionFile for more info on why
		descriptionRaw, err := ReadDescriptionFile(projectFolderPath)
		if err != nil {
			return nil, err
		}

		// Build the list of media filepaths
		mediaFilepaths, err := buildMediaFilepaths(projectFolderPath)
		if err != nil {
			return nil, err
		}

		// Append the new project
		tree = append(tree, ProjectTreeElement{
			ID:             projectFolder.Name(),
			DescriptionRaw: descriptionRaw,
			MediaFilepaths: mediaFilepaths,
		})
	}
	return tree, nil
}

func BuildProjectsTreeScatteredMode(projectsDirectory string) ([]ProjectTreeElement, error) {
	var tree []ProjectTreeElement
	files, err := ioutil.ReadDir(projectsDirectory)
	if err != nil {
		return nil, err
	}
	for _, projectFolder := range files {
		// Not a project folder
		if !projectFolder.IsDir() {
			continue
		}
		portfoliodbDirPath := path.Join(projectsDirectory, projectFolder.Name(), ".portfoliodb")
		portfoliodbDir, err := os.Stat(portfoliodbDirPath)
		// Project has no .portfoliodb file/folder
		if err != nil {
			continue
		}
		// .portfoliodb is not a folder
		if !portfoliodbDir.IsDir() {
			continue
		}
		// Read the description.md file
		// If description is empty, then the project is not portfoliodb-enabled.
		// See ReadDescriptionFile for more info on why
		descriptionRaw, err := ReadDescriptionFile(portfoliodbDirPath)
		if err != nil {
			return nil, err
		}
		if descriptionRaw == "" {
			continue
		}
		// Build the lest of media filepaths
		mediaFilepaths, err := buildMediaFilepaths(portfoliodbDirPath)
		if err != nil {
			return nil, err
		}
		// Append the new project
		tree = append(tree, ProjectTreeElement{
			ID:             projectFolder.Name(),
			DescriptionRaw: descriptionRaw,
			MediaFilepaths: mediaFilepaths,
		})
	}
	return tree, nil
}

// ReadDescriptionFile reads the description.md file in directory.
// Returns an empty string if the file is a directory or does not exist.
func ReadDescriptionFile(directory string) (string, error) {
	descriptionFilepath := path.Join(directory, "description.md")
	if !FileExists(descriptionFilepath) {
		return "", nil
	}
	descriptionFile, err := os.Stat(descriptionFilepath)
	if err != nil {
		return "", err
	}
	if descriptionFile.IsDir() {
		return "", nil
	}
	return ReadFile(descriptionFilepath)
}

func buildMediaFilepaths(at string) ([]string, error) {
	mediaFiles, err := ioutil.ReadDir(at)
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
	return mediaFilepaths, nil
}
