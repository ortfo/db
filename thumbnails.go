package ortfodb

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

var ThumbnailableContentTypes = []string{"image/*", "video/*", "application/pdf"}

func (m Media) Thumbnailable() bool {
	// TODO
	if m.Online {
		return false
	}

	for _, contentTypePattern := range ThumbnailableContentTypes {
		match, err := filepath.Match(contentTypePattern, m.ContentType)
		if err != nil {
			panic(err)
		}
		if match {
			return true
		}
	}
	return false
}

// MakeThumbnail creates a thumbnail on disk of the given media (it is assumed that the given media is an image).
// It returns the path where the thumbnail has been written to.
func (ctx *RunContext) MakeThumbnail(media Media, targetSize uint16, saveTo string) error {
	if media.ContentType == "image/gif" {
		return ctx.makeGifThumbnail(media, targetSize, saveTo)
	}

	if strings.HasPrefix(media.ContentType, "image/") {
		return run("convert", "-resize", fmt.Sprint(targetSize), ctx.AbsolutePathToMedia(media), saveTo)
	}

	if strings.HasPrefix(media.ContentType, "video/") {
		return run("ffmpegthumbnailer", "-i"+ctx.AbsolutePathToMedia(media), "-o"+saveTo, fmt.Sprintf("-s%d", targetSize))
	}

	if media.ContentType == "application/pdf" {
		return ctx.makePdfThumbnail(media, targetSize, saveTo)
	}

	return fmt.Errorf("cannot make a thumbnail for %s: unsupported content type %s", media.Source, media.ContentType)

}

func (ctx *RunContext) makePdfThumbnail(media Media, targetSize uint16, saveTo string) error {
	// If the target extension was not supported, convert from png to the actual target extension
	temporaryPng, err := ioutil.TempFile("", "*.png")
	defer os.Remove(temporaryPng.Name())
	if err != nil {
		return err
	}
	// TODO: (maybe) update media.Dimensions now that we have an image of the PDF though this will only be representative when all pages of the PDF have the same dimensions.
	// pdftoppm *adds* the extension to the end of the filename even if it already has it... smh.
	err = run("pdftoppm", "-singlefile", "-png", ctx.AbsolutePathToMedia(media), strings.TrimSuffix(temporaryPng.Name(), ".png"))
	if err != nil {
		return err
	}
	return run("convert", "-thumbnail", fmt.Sprint(targetSize), temporaryPng.Name(), saveTo)
}

func (ctx *RunContext) makeGifThumbnail(media Media, targetSize uint16, saveTo string) error {
	var dimensionToResize string
	if media.Dimensions.AspectRatio > 1 {
		dimensionToResize = "width"
	} else {
		dimensionToResize = "height"
	}
	source, err := os.Open(ctx.AbsolutePathToMedia(media))
	if err != nil {
		return fmt.Errorf("while opening source media: %w", err)
	}
	defer source.Close()

	tempGif, err := ioutil.TempFile("", "*.gif")
	if err != nil {
		return fmt.Errorf("while creating temporary processed GIF file: %w", err)
	}

	defer tempGif.Close()

	err = runWithStdoutStdin("gifsicle", source, tempGif, "--resize-"+dimensionToResize, fmt.Sprint(targetSize))
	if err != nil {
		return fmt.Errorf("while resizing source GIF: %w", err)
	}

	if strings.HasSuffix(saveTo, ".webp") {
		err = convertGifToWebp(tempGif.Name(), saveTo)
		if err != nil {
			return fmt.Errorf("while converting temporary processed GIF file to webp: %w", err)
		}

	} else {
		dest, err := os.Create(saveTo)
		if err != nil {
			return fmt.Errorf("while creating thumbnail file: %w", err)
		}
		defer dest.Close()
		content, err := os.ReadFile(tempGif.Name())
		if err != nil {
			return fmt.Errorf("while reading temporary processed GIF file: %w", err)
		}
		_, err = dest.Write(content)
		if err != nil {
			return fmt.Errorf("while writing to thumbnail file: %w", err)
		}
	}
	return nil
}

// runWithStdoutStdin runs the given command with the given arguments, setting stdin and stdout to the given readers.
// The returned error contains stdout if the exit code was nonzero.
func runWithStdoutStdin(command string, stdin io.Reader, stdout io.Writer, args ...string) error {
	// Create the proc
	proc := exec.Command(command, args...)

	// Hook up stderr/out to a writer so that we can capture the output
	var stdBuffer bytes.Buffer
	stdWriter := io.MultiWriter(os.Stdout, &stdBuffer)
	proc.Stdout = stdout
	proc.Stderr = stdWriter
	proc.Stdin = stdin

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

func convertGifToWebp(source string, destination string) error {
	return run("gif2webp", "-quiet", source, "-o", destination)
}

// run is like exec.Command(...).Run(...) but the error's message is actually useful (it's not just "exit status n", it has the stdout+stderr)
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

// ComputeOutputThumbnailFilename returns the filename where to save a thumbnail,
// using to the configuration and the given information.
// file name templates are relative to the output database directory.
// Placeholders that will be replaced in the file name template:
//
//	<project id>          the project’s id
//	<media directory>     the value of media.at in the configuration
//	<basename>            the media’s basename (with the extension)
//	<media id>            the media’s id
//	<size>                the current thumbnail size
//	<extension>           the media’s extension
//	<lang>                the current language.
func (ctx *RunContext) ComputeOutputThumbnailFilename(media Media, projectID string, targetSize uint16, lang string) string {
	computed := ctx.Config.MakeThumbnails.FileNameTemplate
	computed = strings.ReplaceAll(computed, "<project id>", projectID)
	computed = strings.ReplaceAll(computed, "<work id>", projectID)
	computed = strings.ReplaceAll(computed, "<media directory>", ctx.Config.Media.At)
	computed = strings.ReplaceAll(computed, "<basename>", path.Base(ctx.AbsolutePathToMedia(media)))
	computed = strings.ReplaceAll(computed, "<media id>", media.ID)
	computed = strings.ReplaceAll(computed, "<size>", fmt.Sprint(targetSize))
	computed = strings.ReplaceAll(computed, "<extension>", strings.Replace(filepath.Ext(ctx.AbsolutePathToMedia(media)), ".", "", 1))
	computed = strings.ReplaceAll(computed, "<lang>", lang)
	return computed
}
