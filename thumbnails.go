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

// StepMakeThumbnails executes the step "make thumbnails" and returns a new metadata object with a new `thumbnails` entry mapping a file's path relative to the database directory to a map mapping a size to a absolute thumbnail filepath
func (ctx *RunContext) StepMakeThumbnails(metadata map[string]interface{}, projectID string, mediae map[string][]Media) (map[string]interface{}, error) {
	alreadyMadeOnes := make([]string, 0)
	madeThumbnails := make(map[string]map[uint16]string)
	for lang, mediae := range mediae {
		for _, media := range mediae {
			madeThumbnails[media.Path] = make(map[uint16]string)
			for _, size := range ctx.Config.MakeThumbnails.Sizes {
				saveTo := path.Join(ctx.DatabaseDirectory, ctx.ComputeOutputThumbnailFilename(media, projectID, size, lang))
				// Don't re-build already-built thumbs
				if StringInSlice(alreadyMadeOnes, saveTo) {
					continue
				}
				// FIXME this is not good, GetBuildMetadata is called in every loop, and it reads a file...
				if !ctx.NeedsRebuiling(saveTo) {
					continue
				}
				// Create potentially missing directories
				os.MkdirAll(filepath.Dir(saveTo), 0777)

				// Make the thumbnail
				err := ctx.makeThumbImage(media, size, saveTo)

				// Handle errors by showing them and setting this source to the empty string
				// Don't return the error, because ending the whole build for one failed thumb would be too much.
				if err != nil {
					fmt.Printf("\n%s\n", err)
					madeThumbnails[media.Path][size] = ""
				} else {
					madeThumbnails[media.Path][size] = ctx.transformSource(saveTo)
				}
			}
		}
	}
	metadata["thumbnails"] = madeThumbnails
	return metadata, nil
}

// makeThumbImage creates a thumbnail on disk of the given media (it is assumed that the given media is an image),
// a target size & the file to save the thumbnail to. Returns the path where the thumbnail has been written.
func (ctx *RunContext) makeThumbImage(media Media, targetSize uint16, saveTo string) error {
	ctx.Status(fmt.Sprintf("Making thumbnail %s", saveTo))
	if strings.HasPrefix(media.ContentType, "image/") {
		return run("convert", "-thumbnail", fmt.Sprint(targetSize), media.AbsolutePath, saveTo)
	}

	if strings.HasPrefix(media.ContentType, "video/") {
		return run("ffmpegthumbnailer", "-i"+media.AbsolutePath, "-o"+saveTo, fmt.Sprintf("-s%d", targetSize))
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
			// TODO: (maybe) update media.Dimensions now that we have an image of the PDF though this will only be representative when all pages of the PDF have the same dimensions.
			// FIXME: PDF thumbnails are squares instead of respecting the page's aspect ratio.
			// pdftoppm *adds* the extension to the end of the filename even if it already has it... smh.
			err = run("pdftoppm", "-singlefile", "-scale-to", fmt.Sprint(targetSize), "-sz", fmt.Sprint(targetSize), "-png", media.AbsolutePath, strings.TrimSuffix(temporaryPng.Name(), ".png"))
			if err != nil {
				return err
			}
			return run("convert", temporaryPng.Name(), saveTo)
		} else {
			// Else, just use the right flag “-{targetExtension}”
			return run("pdftoppm", "-singlefile", "-scale-to", fmt.Sprint(targetSize), "-sz", fmt.Sprint(targetSize), "-"+targetExtension, saveTo)
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
// It uses media.Source because we might want to compute thumbnails of online media in the future
// Placeholders that will be replaced in the file name template:
//
// * <project id> - the project's id
// * <parent> - the current media's directory
// * <basename> - the media's basename (with the extension)
// * <media id> - the media's id
// * <size> - the current thumbnail size
// * <extension> - the media's extension
// * <lang> - the current language
func (ctx *RunContext) ComputeOutputThumbnailFilename(media Media, projectID string, targetSize uint16, lang string) string {
	computed := ctx.Config.MakeThumbnails.FileNameTemplate
	computed = strings.ReplaceAll(computed, "<project id>", projectID)
	computed = strings.ReplaceAll(computed, "<parent>", filepath.Dir(media.Path)) // FIXME: depends on `replace media sources` removing the /home/ewen/projects/portfolio
	computed = strings.ReplaceAll(computed, "<basename>", path.Base(media.AbsolutePath))
	computed = strings.ReplaceAll(computed, "<media id>", FilepathBaseNoExt(media.AbsolutePath))
	computed = strings.ReplaceAll(computed, "<size>", fmt.Sprint(targetSize))
	computed = strings.ReplaceAll(computed, "<extension>", strings.Replace(filepath.Ext(media.AbsolutePath), ".", "", 1))
	computed = strings.ReplaceAll(computed, "<lang>", lang)
	return computed
}
