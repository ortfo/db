package ortfodb

// Functions to analyze media files.
// Used to go from a ParsedDescription struct to a Work struct.

import (
	"errors"
	"fmt"
	"image"
	"time"

	// Supported formats
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "github.com/hullerob/go.farbfeld"
	_ "github.com/jbuchbinder/gopnm" // PBM, PGM and PPM
	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/ccitt"
	_ "golang.org/x/image/riff"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/vp8"
	_ "golang.org/x/image/vp8l"
	_ "golang.org/x/image/webp"

	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gabriel-vasile/mimetype"
	// "github.com/gen2brain/go-fitz" // FIXME requires cgo, which goreleaser has a hard time with
	"github.com/lafriks/go-svg"
	"github.com/metal3d/go-slugify"
	recurcopy "github.com/plus3it/gorecurcopy"
	ffmpeg "github.com/ssttevee/go-ffmpeg"
	"github.com/tcolgate/mp3"
)

func (p FilePathInsideMediaRoot) Absolute(ctx *RunContext) string {
	result, _ := filepath.Abs(filepath.Join(ctx.Config.Media.At, string(p)))
	return result
}

// RelativeToMediaRoot returns the path to the media file relative to the media root.
//
//	input:   ./[work id]/[scattered mode folder]/[file path]
//	                                             -----------
//	                                             part of the path
//	output:  ./[work id]/[scattered mode folder]/[file path]
//	         -----------------------------------------------
//	         part of the path
func (p FilePathInsidePortfolioFolder) RelativeToMediaRoot(ctx *RunContext, workID string) FilePathInsideMediaRoot {
	return FilePathInsideMediaRoot(filepath.Join(workID, ctx.Config.ScatteredModeFolder, string(p)))
}

// ImageDimensions represents metadata about a media as it's extracted from its file.
type ImageDimensions struct {
	Width       int     `json:"width"`       // Width in pixels
	Height      int     `json:"height"`      // Height in pixels
	AspectRatio float32 `json:"aspectRatio"` // width / height
}

// Media represents a media object inserted in the work object's media array.
type Media struct {
	Alt               string                        `json:"alt"`
	Caption           string                        `json:"caption"`
	RelativeSource    FilePathInsidePortfolioFolder `json:"relativeSource"`
	DistSource        FilePathInsideMediaRoot       `json:"distSource"`
	ContentType       string                        `json:"contentType"`
	Size              int                           `json:"size"` // in bytes
	Dimensions        ImageDimensions               `json:"dimensions"`
	Online            bool                          `json:"online"`
	Duration          float64                       `json:"duration"` // in seconds
	HasSound          bool                          `json:"hasSound"`
	Colors            ColorPalette                  `json:"colors"`
	Thumbnails        ThumbnailsMap                 `json:"thumbnails"`
	ThumbnailsBuiltAt time.Time                     `json:"thumbnailsBuiltAt"`
	Attributes        MediaAttributes               `json:"attributes"`
	Analyzed          bool                          `json:"analyzed"` // whether the media has been analyzed
	// Hash of the media file, used for caching purposes. Could also serve as an integrity check.
	// The value is the MD5 hash, base64-encoded.
	Hash string `json:"hash"`
}

// GetImageDimensions returns an ImageDimensions object, given a pointer to a file.
func GetImageDimensions(file *os.File) (ImageDimensions, error) {
	img, _, err := image.Decode(file)
	if err != nil {
		return ImageDimensions{}, err
	}
	// Get height & width
	height := img.Bounds().Dy()
	width := img.Bounds().Dx()
	// Get aspect ratio
	ratio := float32(width) / float32(height)
	return ImageDimensions{width, height, ratio}, nil
}

func removeUnit(s string) string {
	return strings.TrimSuffix(s, "px")
}

// GetSVGDimensions returns an ImageDimensions object, given a pointer to a SVG file.
// If neither viewBox nor width & height attributes are set, the resulting dimensions will be 0x0.
func GetSVGDimensions(file *os.File) (ImageDimensions, error) {
	svg, err := svg.Parse(file, svg.IgnoreErrorMode)
	if err != nil {
		return ImageDimensions{}, fmt.Errorf("while parsing SVG file: %w", err)
	}

	var height, width float64
	if svg.Width != "" && svg.Height != "" {
		height, err = strconv.ParseFloat(removeUnit(svg.Height), 32)
		if err != nil {
			return ImageDimensions{}, fmt.Errorf("cannot parse SVG height attribute as a number: %w", err)
		}
		width, err = strconv.ParseFloat(removeUnit(svg.Width), 32)
		if err != nil {
			return ImageDimensions{}, fmt.Errorf("cannot parse SVG width attribute as a number: %w", err)
		}
	} else if svg.ViewBox.H != 0 && svg.ViewBox.W != 0 {
		height = float64(svg.ViewBox.H)
		width = float64(svg.ViewBox.W)
	} else {
		return ImageDimensions{}, fmt.Errorf("cannot determine dimensions of SVG file")
	}

	return ImageDimensions{
		Width:       int(width),
		Height:      int(height),
		AspectRatio: float32(width) / float32(height),
	}, nil
}

// PathToWorkFolder returns the path to the work's folder, including the .portfoliodb part if --scattered.
func (ctx *RunContext) PathToWorkFolder(workID string) string {
	path := filepath.Join(ctx.DatabaseDirectory, workID)
	if ctx.Flags.Scattered {
		path = filepath.Join(path, ctx.Config.ScatteredModeFolder)
	}
	return path
}

// AnalyzeMediaFile analyzes the file at its absolute filepath filename and returns a Media struct, merging the analysis' results with information from the matching MediaEmbedDeclaration.
// TODO prevent duplicate analysis of the same file in the current session even when file was never analyzed on previous runs of the command
func (ctx *RunContext) AnalyzeMediaFile(workID string, embedDeclaration Media) (usedCache bool, analyzedMedia Media, anchor string, err error) {
	defer TimeTrack(time.Now(), "AnalyzeMediaFile", workID, embedDeclaration.RelativeSource)
	LogDebug("Analyzing media %#v", embedDeclaration)

	// Compute absolute filepath to media
	var filename string
	if !filepath.IsAbs(string(embedDeclaration.RelativeSource)) {
		filename, _ = filepath.Abs(filepath.Join(ctx.PathToWorkFolder(workID), string(embedDeclaration.RelativeSource)))
	} else {
		filename = string(embedDeclaration.RelativeSource)
	}
	anchor = slugify.Marshal(filepathBaseNoExt(filename), true)
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()
	fileInfo, err := file.Stat()
	if err != nil {
		return
	}

	var contentType string
	var contentHash string

	if fileInfo.IsDir() {
		contentType = "directory"
	} else {
		var cachedAnalysis Media
		contentHash, usedCache, cachedAnalysis, err = ctx.UseMediaCache(filename, embedDeclaration, workID)
		if err != nil {
			return false, Media{}, "", fmt.Errorf("while evaluating whether to use cache for media %s: %w", filename, err)
		}

		if usedCache && cachedAnalysis.ContentType != "" {
			LogDebug("Reusing cached analysis %#v", cachedAnalysis)
			return true, cachedAnalysis, anchor, nil
		} else if usedCache {
			LogDebug("UseMediaCache tells me to use cache for %s, but the cached analysis has no content type. Will reanalyze.", filename)
		}

		ctx.Status(workID, PhaseMediaAnalysis, string(embedDeclaration.RelativeSource))
		mimeType, err := mimetype.DetectFile(filename)
		if err != nil {
			contentType = "application/octet-stream"
		} else {
			contentType = mimeType.String()
		}
	}

	isAudio := strings.HasPrefix(contentType, "audio/")
	isVideo := strings.HasPrefix(contentType, "video/")
	isImage := strings.HasPrefix(contentType, "image/")
	isPDF := contentType == "application/pdf"

	var dimensions ImageDimensions
	var duration uint
	var hasSound bool
	var colors ColorPalette

	if isImage {
		if contentType == "image/svg" || contentType == "image/svg+xml" {
			dimensions, err = GetSVGDimensions(file)
		} else {
			dimensions, err = GetImageDimensions(file)
		}
		if err != nil {
			return
		}
		if ctx.Config.ExtractColors.Enabled {
			if canExtractColors(contentType) {
				LogDebug("Extracting colors from %s", filename)
				colors, err = ExtractColors(filename)
				if err != nil {
					DisplayError("Could not extract colors from %s", err, filename)
					err = nil
				}
				LogDebug("Colors extracted from %s: %#v", filename, colors)
			} else {
				LogDebug("Not extracting colors from %s: unsupported content type", filename)
			}
		}
	}

	if isVideo {
		dimensions, duration, hasSound, err = AnalyzeVideo(filename)
		if err != nil {
			return
		}
		LogDebug("Video analyzed: dimensions=%#v, duration=%v, hasSound=%v", dimensions, duration, hasSound)
	}

	if isAudio {
		duration = AnalyzeAudio(file)
		hasSound = true
		LogDebug("Audio analyzed: duration=%v", duration)
	}

	if isPDF {
		LogWarning("PDF analysis is disabled")
		// dimensions, duration, err = AnalyzePDF(filename)
		// if err != nil {
		// 	return
		// }
		// LogDebug("PDF analyzed: dimensions=%#v, duration=%v", dimensions, duration)
	}

	analyzedMedia = Media{
		Alt:            embedDeclaration.Alt,
		Caption:        embedDeclaration.Caption,
		RelativeSource: embedDeclaration.RelativeSource,
		DistSource:     FilePathInsideMediaRoot(embedDeclaration.RelativeSource.RelativeToMediaRoot(ctx, workID)),
		Attributes:     embedDeclaration.Attributes,
		ContentType:    contentType,
		Dimensions:     dimensions,
		Duration:       float64(duration),
		Size:           int(fileInfo.Size()),
		HasSound:       hasSound,
		Colors:         colors,
		Analyzed:       true,
		Hash:           contentHash,
	}
	LogDebug("Analyzed to %#v (no cache used)", analyzedMedia)
	return
}

func (ctx *RunContext) UseMediaCache(filename string, embedDeclaration Media, workID string) (newHash string, used bool, media Media, err error) {
	if ctx.Flags.NoCache {
		LogDebug("--no-cache is set, eagerly computing hash")
		newHash, err = hashFile(filename)
		if err != nil {
			err = fmt.Errorf("while computing hash of media: %w", err)
			return
		}
		return
	}

	LogDebug("checking mtime of %s before trying to compute hashes", filename)
	stat, err := os.Stat(filename)
	if err != nil {
		err = fmt.Errorf("could not check modification times of %s: %w", filename, err)
		return
	}

	var oldWork AnalyzedWork
	var oldMedia Media
	var oldMediaFound bool

	if oldMedia, oldWork, oldMediaFound = ctx.PreviouslyBuiltMedia(workID, embedDeclaration); oldMediaFound {
		if embedDeclaration.Hash == "" {
			LogDebug("media %s in old database has no hash stored, hash will be computed.", filename)
		} else if oldWork.BuiltAt.After(stat.ModTime()) {
			LogDebug("mtime cache strategy: not recomputing hash of %s because it was last modified before the previous build (file modified at %s, previous build at %s) , using cached analysis", filename, stat.ModTime(), oldWork.BuiltAt)
			LogDebug("cache hit by modtime for %s: using cache from embed decl %#v", filename, embedDeclaration)
			return embedDeclaration.Hash, true, oldMedia, nil
		} else {
			LogDebug("file mtime of %s is newer than previous build (file modified at %s, previous build at %s), computing hash", filename, stat.ModTime(), ctx.BuildMetadata.PreviousBuildDate)
		}
	} else {
		LogDebug("mtime cache strategy: media %s of %s not found in previous database, will compute hash", filename, workID)
	}

	newHash, err = hashFile(filename)
	if err != nil {
		err = fmt.Errorf("while computing hash of media: %w", err)
		return
	}

	if oldMediaFound {
		if oldMedia.Hash == newHash {
			LogDebug("cache hit by hash for %s: using cache from embed decl %#v", filename, embedDeclaration)
			return newHash, true, oldMedia, nil
		}
		LogDebug("Cache miss for %s: old content hash %q is different from %q", filename, oldMedia.Hash, newHash)
		return newHash, false, embedDeclaration, nil
	}

	LogDebug("Cache miss for %s: media not found in previous database build", filename)
	return

}

// AnalyzeAudio takes in an os.File and returns the duration of the audio file in seconds. If any error occurs the duration will be 0.
func AnalyzeAudio(file *os.File) uint {
	var duration uint
	decoder := mp3.NewDecoder(file)
	skipped := 0
	var frame mp3.Frame
	for {
		err := decoder.Decode(&frame, &skipped)
		if err != nil {
			break
		} else {
			duration += uint(frame.Duration().Seconds())
		}
	}
	return duration
}

// AnalyzePDF returns an ImageDimensions struct for the first page of the PDF file at filename. It also returns the number of pages.
func AnalyzePDF(filename string) (dimensions ImageDimensions, pagesCount uint, err error) {
	return ImageDimensions{}, 0, errors.New("PDF analysis is disabled")
	/* fitz requires cgo, which goreleaser has a hard time with.
	document, err := fitz.New(filename)
	if err != nil {
		return dimensions, pagesCount, fmt.Errorf("while opening PDF: %w", err)
	}

	defer document.Close()

	firstPage, err := document.Image(1)
	if err != nil {
		return dimensions, pagesCount, fmt.Errorf("while getting an image of the PDF's first page: %w", err)
	}

	width := firstPage.Bounds().Size().X
	height := firstPage.Bounds().Size().Y

	return ImageDimensions{
		Width:       int(width),
		Height:      int(height),
		AspectRatio: float32(width) / float32(height),
	}, uint(document.NumPage()), nil
	*/
}

// AnalyzeVideo returns an ImageDimensions struct with the video's height, width and aspect ratio and a duration in seconds.
func AnalyzeVideo(filename string) (dimensions ImageDimensions, duration uint, hasSound bool, err error) {
	probe, err := ffmpeg.DefaultConfiguration()
	if err != nil {
		return
	}
	_, video, err := probe.Probe(filename)
	if err != nil {
		return
	}
	for _, stream := range video.Streams {
		if stream.CodecType == "audio" {
			hasSound = true
		}
		if stream.CodecType == "video" {
			dimensions = ImageDimensions{
				Height:      stream.Height,
				Width:       stream.Width,
				AspectRatio: float32(stream.Width) / float32(stream.Height),
			}
		}
	}
	durationFloat, parseErr := strconv.ParseFloat(video.Format.Duration, 64)
	if parseErr != nil {
		err = fmt.Errorf("couldn't convert media duration %#v to number: %w", video.Format.Duration, parseErr)
		return
	}
	duration = uint(durationFloat)
	return
}

func (ctx *RunContext) HandleMedia(workID string, blockID string, embedDeclaration Media, language string) (media Media, anchor string, usedCache bool, err error) {
	defer TimeTrack(time.Now(), "HandleMedia", workID, embedDeclaration.RelativeSource)
	usedCache, media, anchor, err = ctx.AnalyzeMediaFile(workID, embedDeclaration)
	if err != nil {
		err = fmt.Errorf("while analyzing media: %w", err)
		return
	}

	// Copy over
	if ctx.Config.Media.At == "" {
		err = errors.New("please specify a destination for the media files in the configuration file (set media.at)")
		return
	}

	absolutePathSource := media.RelativeSource.Absolute(ctx, workID)
	absolutePathDestination := media.DistSource.Absolute(ctx)

	copyingStepStart := time.Now()
	skipCopy := usedCache && fileExists(absolutePathDestination)
	if skipCopy {
		LogDebug("Skipping media copy for %s because it already exists", absolutePathDestination)
	}
	if absolutePathDestination != absolutePathSource && !skipCopy {
		err = os.MkdirAll(path.Dir(absolutePathDestination), 0o755)
		if err != nil {
			err = fmt.Errorf("could not create output directory for %s: %w", absolutePathSource, err)
			return
		}
		if media.ContentType == "directory" {
			err = recurcopy.CopyDirectory(absolutePathSource, absolutePathDestination)
		} else {
			// content, err = os.ReadFile(absolutePathSource)
			// if err != nil {
			// 	err = fmt.Errorf("could not read file %s: %w", absolutePathSource, err)
			// 	return
			// }
			// err = os.WriteFile(absolutePathDestination, content, 0777)
			err = copyFile(absolutePathSource, absolutePathDestination)
		}

		if err != nil {
			err = fmt.Errorf("while copying media over: %w", err)
			return
		}
	}
	TimeTrack(copyingStepStart, "HandleMedia > copy to dist", media.RelativeSource, media.DistSource)

	thumbnailsStepStart := time.Now()
	// Make thumbnail
	if media.Thumbnailable() && ctx.Config.MakeThumbnails.Enabled {
		media.Thumbnails = make(map[int]FilePathInsideMediaRoot)
		type result struct {
			size    int
			err     error
			skipped bool
		}
		builtSizes := 0

		results := make(chan result)

		for i, sizesToDo := range chunkSlice(ctx.Config.MakeThumbnails.Sizes, ctx.thumbnailersPerWork) {
			go func(i int, sizesToDo []int, results chan result) {
				for _, size := range sizesToDo {
					LogDebug("Making thumbnail @%d for %s#%s", size, media.RelativeSource, blockID)
					saveTo := ctx.ComputeOutputThumbnailFilename(media, blockID, workID, size, language)

					if _, err := os.Stat(string(saveTo.Absolute(ctx))); err == nil && usedCache {
						LogDebug("Skipping thumbnail creation @%d for %s#%s because it already exists", size, media.RelativeSource, blockID)
						results <- result{size: size, skipped: true}
						continue
					}

					// Create potentially missing directories
					os.MkdirAll(filepath.Dir(saveTo.Absolute(ctx)), 0777)

					ctx.Status(workID, PhaseThumbnails, string(media.RelativeSource), fmt.Sprintf("%dpx", size))

					// Make the thumbnail
					err := ctx.MakeThumbnail(media, size, saveTo.Absolute(ctx))
					if err != nil {
						results <- result{err: fmt.Errorf("while making thumbnail @%d for %s: %w", size, workID, err)}
						continue
					}
					LogDebug("Made thumbnail %s", saveTo)
					results <- result{size: size}
				}
			}(i, sizesToDo, results)
		}

		for result := range results {
			if result.err != nil {
				return media, anchor, usedCache, result.err
			}
			if !result.skipped {
				media.Thumbnails[result.size] = ctx.ComputeOutputThumbnailFilename(media, blockID, workID, result.size, language)
				media.ThumbnailsBuiltAt = time.Now()
			}
			builtSizes++

			if builtSizes >= len(ctx.Config.MakeThumbnails.Sizes) {
				close(results)
			}
		}
	}
	TimeTrack(thumbnailsStepStart, "HandleMedia > thumbnails", media.RelativeSource)

	return
}
