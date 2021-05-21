// thumbnails provides the StepMakeThumbnails step.
// It assumes that several commands are available to the shell:
// magick (tried using a library but it made my computer freeze while high on RAM),
// ffmpegthumbnailer,
// pdftoppm

package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

//TODO: convert GIFs from `online: True` sources (YouTube, Dailymotion, Vimeo, you name it.). Might want to look at <https://github.com/hunterlong/gifs>

// StepMakeThumbnails executes the step "make thumbnails" and returns a new metadata object with a new `thumbnails` entry mapping a file to a map mapping a size to a thumbnail filepath
func StepMakeThumbnails(metadata map[string]interface{}, project ProjectTreeElement, databaseDirectory string, mediae map[string][]Media, config Configuration) (map[string]interface{}, error) {
	alreadyMadeOnes := make([]string, 0)
	madeThumbnails := make(map[string]map[uint16]string)
	for lang, mediae := range mediae {
		for _, media := range mediae {
			// matches, err := filepath.Match(config.MakeThumbnails.InputFile, media.Source)
			// if err != nil || !matches || config.MakeThumbnails.InputFile == "" {
			// 	continue
			// }
			madeThumbnails[transformSource(media.Source, config)] = make(map[uint16]string)
			for _, size := range config.MakeThumbnails.Sizes {
				saveTo := path.Join(databaseDirectory, ComputeOutputThumbnailFilename(config, media, project, size, lang))
				if StringInSlice(alreadyMadeOnes, saveTo) {
					continue
				}
				if media.Dimensions.AspectRatio == 0.0 {
					continue
				}
				// FIXME this is not good, GetBuildMetadata is called in every loop, and it reads a file...
				if !NeedsRebuiling(saveTo, config) {
					continue
				}
				err := makeThumbImage(media, size, saveTo, databaseDirectory)
				// Create potentially missing directories
				os.MkdirAll(filepath.Dir(saveTo), 0777)
				if err != nil {
					return nil, err
				}
				madeThumbnails[transformSource(media.Source, config)][size] = transformSource(saveTo, config)
			}
		}
	}
	metadata["thumbnails"] = madeThumbnails
	return metadata, nil
}

// makeThumbImage creates a thumbnail on disk of the given media (it is assumed that the given media is an image),
// a target size & the file to save the thumbnail to. Returns the path where the thumbnail has been written.
func makeThumbImage(media Media, targetSize uint16, saveTo string, databaseDirectory string) error {
	mediaAbsoluteSource := path.Join(databaseDirectory, media.Source)

	if strings.HasPrefix(media.ContentType, "image/") {
		return run("convert", "-thumbnail", fmt.Sprint(targetSize), mediaAbsoluteSource, saveTo)
	}

	if strings.HasPrefix(media.ContentType, "video/") {
		return run("ffmpegthumbnailer", "-i"+mediaAbsoluteSource, "-o"+saveTo, fmt.Sprintf("-s%d", targetSize))
	}

	if media.ContentType == "application/pdf" {
		supportedExtensions := "png jpeg tiff"
		targetExtension := filepath.Ext(saveTo)
		if targetExtension == "jpg" {
			// jpg is jpeg
			targetExtension = "jpeg"
		}
		if !strings.Contains(supportedExtensions, targetExtension) {
			// If the target extension was not supported, convert from png to the actual target extension
			temporaryPng, err := ioutil.TempFile("", "*.png")
			defer os.Remove(temporaryPng.Name())
			if err != nil {
				return err
			}
			err = run("pdftoppm", "-singlefile", "-scale-to-x", fmt.Sprint(targetSize), "-png", temporaryPng.Name())
			if err != nil {
				return err
			}
			return run("convert", temporaryPng.Name(), saveTo)
		} else {
			// Else, just use the right flag “-{targetExtension}”
			return run("pdftoppm", "-singlefile", "-scale-to-x", fmt.Sprint(targetSize), "-"+targetExtension, saveTo)
		}
	}

	return fmt.Errorf("cannot make a thumbnail for %s: unsupported content type %s", media.Source, media.ContentType)

}

// run is like exec.Command(...).Run(...) but the error's message is actually useful (it's not just "exit status n")
func run(command string, args ...string) error {
	// Create the proc
	proc := exec.Command(command, args...)

	// Hook up stderr/out to a writer so that we can capture the output
	var stdBuffer bytes.Buffer
	stdWriter := io.MultiWriter(os.Stdout, &stdBuffer)
	proc.Stdout = stdWriter
	proc.Stderr = stdWriter

	// Run the proc
	err := proc.Run()

	// Handle errors
	if err != nil {
		switch e := err.(type) {
		case *exec.ExitError:
			return fmt.Errorf("while running %s: exited with %d: %s", strings.Join(proc.Args, " "), e.ExitCode(), stdBuffer.String())

		default:
			return fmt.Errorf("while running %s: %s", strings.Join(proc.Args, " "), err.Error())
		}
	}
	return nil
}

// ComputeOutputThumbnailFilename returns the filename where to save a thumbnail
// according to the configuration and the given information.
// file name templates are relative to the output database directory.
// Placeholders that will be replaced in the file name template:
//
// * <project id> - the project's id
// * <parent> - the current media's directory
// * <basename> - the media's basename (with the extension)
// * <media id> - the media's id
// * <size> - the current thumbnail size
// * <extension> - the media's extension
// * <lang> - the current language
func ComputeOutputThumbnailFilename(config Configuration, media Media, project ProjectTreeElement, targetSize uint16, lang string) string {
	computed := config.MakeThumbnails.FileNameTemplate
	computed = strings.ReplaceAll(computed, "<project id>", project.ID)
	computed = strings.ReplaceAll(computed, "<parent>", filepath.Dir(media.Source))
	computed = strings.ReplaceAll(computed, "<basename>", path.Base(media.Source))
	computed = strings.ReplaceAll(computed, "<media id>", FilepathBaseNoExt(media.Source))
	computed = strings.ReplaceAll(computed, "<size>", fmt.Sprint(targetSize))
	computed = strings.ReplaceAll(computed, "<extension>", strings.Replace(filepath.Ext(media.Source), ".", "", 1))
	computed = strings.ReplaceAll(computed, "<lang>", lang)
	return computed
}
