package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"
)

const (
	smashingMagazineURLFirstPage = "https://www.smashingmagazine.com/category/wallpapers/"
	smashingMagazineURLTpl       = "https://www.smashingmagazine.com/category/wallpapers/page/%d/"
	wallpaperPageURLPartTpl      = "%s-%02d"                      // {month}-{year}, e.g.: october-2016
	maxScapedPages               = 13                             // fairly chosed by random
	wallpaperURLPatternTpl       = "%s.*%s\\.(jpg|jpeg|png|gif)$" // {cal or nocal}-{resolution}.{extension}
)

// Helper function to detect user home dir on windows & linux
func userHomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}

// Helper function to pull the href attribute from a Token
func getHref(t html.Token) (ok bool, href string) {
	// Iterate over all of the Token's attributes until we find an "href"
	for _, a := range t.Attr {
		if a.Key == "href" {
			href = a.Val
			ok = true
		}
	}

	// "bare" return will return the variables (ok, href) as defined in
	// the function definition
	return
}

// Helper function to download file from URL to provided directory
func downloadFromURL(url string, directory string) {
	tokens := strings.Split(url, "/")
	fileName := path.Join(directory, tokens[len(tokens)-1])
	log.Println("Downloading", url, "to", fileName)

	// TODO: check file existence first with io.IsExist
	output, err := os.Create(fileName)
	if err != nil {
		log.Println("Error while creating", fileName, "-", err)
		return
	}
	defer output.Close()

	response, err := http.Get(url)
	if err != nil {
		log.Println("Error while downloading", url, "-", err)
		return
	}
	defer response.Body.Close()

	n, err := io.Copy(output, response.Body)
	if err != nil {
		log.Println("Error while downloading", url, "-", err)
		return
	}

	log.Println(n, "bytes downloaded.")
}

// Find all urls in html matched provided regexp
func findURLsInPage(url string, pattern string) (urls []string) {
	resp, err := http.Get(url)
	if err != nil {
		// handle error
		log.Fatal(err)
		return
	}
	defer resp.Body.Close()

	z := html.NewTokenizer(resp.Body)
	for {
		tt := z.Next()

		switch {
		case tt == html.ErrorToken:
			// End of the document, we're done
			return
		case tt == html.StartTagToken:
			t := z.Token()

			// Check if the token is an <a> tag
			isAnchor := t.Data == "a"
			if !isAnchor {
				continue
			}

			// Extract the href value, if there is one
			ok, href := getHref(t)
			if !ok {
				continue
			}
			// Make sure the url begines in http**
			neededURL, _ := regexp.MatchString(pattern, href)
			// log.Println(neededURL, err, pattern, href)
			if neededURL {
				urls = append(urls, href)
			}
		}
	}
}

func findWallpaperURL(year int, monthName string) (url string) {
	wallpaperURLPart := fmt.Sprintf(wallpaperPageURLPartTpl, monthName, year)

	for i := 1; i <= maxScapedPages; i++ {
		smashingMagazineURL := fmt.Sprintf(smashingMagazineURLTpl, i)
		if i == 1 {
			smashingMagazineURL = smashingMagazineURLFirstPage
		}
		urls := findURLsInPage(smashingMagazineURL, wallpaperURLPart)
		if len(urls) > 0 {
			return urls[0]
		}
		break
	}
	return
}

func main() {
	var (
		year       int
		month      int
		resolution string
		nocal      bool
	)

	// TODO: check path
	basePicturesDirectory := path.Join(userHomeDir(), "Pictures", "Smashing-Wallpapers")

	// default params for month & year
	currentYear, currentMonth, _ := time.Now().UTC().Date()

	// CLI flags
	flag.IntVar(&year, "year", currentYear, "Specify year, default to current")
	flag.IntVar(&year, "y", currentYear, "Specify year, default to current (shorthand)")
	flag.IntVar(&month, "month", int(currentMonth), "Specify month, default to current")
	flag.IntVar(&month, "m", int(currentMonth), "Specify month, default to current (shorthand)")
	flag.StringVar(&resolution, "resolution", "1920x1080", "Specify wallpaper resolution")
	flag.StringVar(&resolution, "r", "1920x1080", "Specify wallpaper resolution (shorthand)")
	flag.BoolVar(&nocal, "nocal", false, "Download wallpapers without calendars")
	flag.Parse()

	if year < 2012 {
		log.Println("Year need to be greater than 2012")
		return
	}
	if month < 1 && month > 12 {
		log.Println("Month valid range is 1..12")
	}
	// TODO: validate resolution

	log.Println("Start program with params: year", year, "month", month, "resolution", resolution, "nocal", nocal)
	monthName := strings.ToLower(time.Month(month).String())

	log.Println("Start to find wallpaper url")

	wallpaperURL := findWallpaperURL(year, monthName)
	wallpaperURL = fmt.Sprintf("https://www.smashingmagazine.com%s", wallpaperURL)

	log.Println("Found wallpaper url", wallpaperURL)

	subPath := fmt.Sprintf("%d.%02d", year, month)
	picturesDirectory := path.Join(basePicturesDirectory, subPath)
	os.MkdirAll(picturesDirectory, 0777)
	log.Println("Will download to directory", picturesDirectory)

	wallpaperURLPattern := fmt.Sprintf(wallpaperURLPatternTpl, "[^o]cal", resolution)
	wallpapersToDownload := findURLsInPage(wallpaperURL, wallpaperURLPattern)
	wg := new(sync.WaitGroup)
	wg.Add(len(wallpapersToDownload))
	for i := 0; i < len(wallpapersToDownload); i++ {
		go func(i int) {
			downloadFromURL(wallpapersToDownload[i], picturesDirectory)
			wg.Done()
		}(i)
	}
	wg.Wait()

	log.Println("Download completed")
}
