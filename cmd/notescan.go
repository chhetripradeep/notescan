package main

import (
	"flag"
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"log"
	"os"
	"runtime/pprof"
	"strings"
	"sync"

	"github.com/chhetripradeep/notescan"
)

var (
	samplingRate = flag.Float64("samplingRate", 0.05, "The background color and the foreground color are selected according to the selected number.")
	shift = flag.Int("shift", 2, "Number of shifts when compressing pixels.")

	brightness = flag.Float64("brightness", 0.35, "Brightness distance when selecting foreground color.")
	saturation = flag.Float64("saturation", 0.25, "Saturation distance when selecting foreground color.")

	foregroundNum = flag.Int("foregroundNum", 6, "Specify the number chosen for the foreground.")
	kMeansIterations = flag.Int("kMeansIterations", 40, "Number of iterations for K-Means algorithm.")

	profileType = flag.String("profileType", "", "Type of profiling to do.")
	suffix = flag.String("suffix", "_processed", "Suffix in the output filename.")
	gif = flag.Bool("gif", false, "Should the output file format be gif.")
)

func Usage() {
	fmt.Println("You can specify multiple input files to process as arguments.")
	flag.Usage()
}

func main() {

	// parse the flags
	flag.Parse()

	// set options from flags
	opt := notescan.Option {
		SamplingRate: *samplingRate,
		Shift: *shift,
		Brightness: *brightness,
		Saturation: *saturation,
		ForegroundNum: *foregroundNum,
		KMeansIterations: *kMeansIterations,
	}

	// process input filenames
	files := flag.Args()
	if files == nil || len(files) == 0 {
		Usage()
		return
	}

	// process each input file asynchronously
	wg := sync.WaitGroup{}
	for _, f := range files {
		wg.Add(1)
		go func(file string) {
			err := run(file, &opt)
			if err != nil {
				fmt.Printf("[%v]\n", err)
			}
			wg.Done()
		}(f)
	}
	wg.Wait()

	return
}

// perform file conversion
func run(f string, opt *notescan.Option) error {
	log.Printf("Shrink: [%s]\n", f)

	// load input image
	in, err := loadImage(f)
	if err != nil {
		return err
	}

	// compress image
	shrink, err := notescan.Shrink(in, opt)
	if err != nil {
		return err
	}

	// output filename
	output := ""
	extension := ".png"
	if *gif {
		extension = ".gif"
	}

	idx := strings.LastIndex(f, ".")
	if idx == -1 {
		output = f + *suffix + extension
	} else {
		output = f[:idx] + *suffix + extension
	}

	// create output image
	if *gif {
		err = notescan.OutputGIF(output, shrink)
	} else {
		err = notescan.OutputPNG(output, shrink)
	}

	if err == nil {
		log.Printf("Generated: [%s]\n", output)
	}

	return err
}

// load input image
func loadImage(f string) (image.Image, error) {
	file, err := os.Open(f)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	return img, nil
}

// profiling
type profile struct {
	file *os.File
	err error
}

// start profiling for performance optimization
func startProfile(f string) *profile {
	log.Println("Profile Start: " + f)

	prof := profile{}
	file, err := os.Create(f)
	if err != nil {
		prof.err = err
	} else {
		err = pprof.StartCPUProfile(file)
		if err == nil {
			prof.file = file
		} else {
			prof.err = err
			defer file.Close()
		}
	}

	return &prof
}

// stop profiling for performance optimization
func (prof profile) stop() {
	log.Println("Profile Stop")
	if prof.err == nil {
		pprof.StopCPUProfile()
		prof.file.Close()
	} else {
		log.Println(prof.err)
	}
}
