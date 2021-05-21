package main

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"gopkg.in/gographics/imagick.v3/imagick"
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
	imagick.Initialize()
	defer imagick.Terminate()
	mediaAbsoluteSource := path.Join(databaseDirectory, media.Source)

	wand := imagick.NewMagickWand()
	err := wand.ReadImage(mediaAbsoluteSource)
	if err != nil {
		return err
	}

	// Two cases depending on the orientation of the image
	var scaledWidth, scaledHeight uint
	if media.Dimensions.AspectRatio >= 1 {
		scaledWidth = uint(targetSize)
		scaledHeight = uint(1 / float32(media.Dimensions.AspectRatio) * float32(scaledWidth))
	} else {
		scaledHeight = uint(targetSize)
		scaledWidth = uint(float32(media.Dimensions.AspectRatio) * float32(scaledHeight))
	}

	err = wand.AdaptiveResizeImage(scaledWidth, scaledHeight)
	if err != nil {
		return err
	}
	err = wand.SetImageCompressionQuality(65)
	if err != nil {
		return err
	}
	err = wand.WriteImage(saveTo)
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
