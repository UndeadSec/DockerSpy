package main

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type IgnoreExtensions struct {
	Extensions []string `json:"extensions"`
}

type SearchResult struct {
	NumResults int    `json:"count"`
	Next       string `json:"next"`
	Results    []struct {
		Name        string `json:"repo_name"`
		Description string `json:"short_description"`
		PullCount   int    `json:"pull_count"`
		StarCount   int    `json:"star_count"`
		IsOfficial  bool   `json:"is_official"`
	} `json:"results"`
}

type TagsResult struct {
	Count    int    `json:"count"`
	Next     string `json:"next"`
	Previous string `json:"previous"`
	Results  []struct {
		Name string `json:"name"`
	} `json:"results"`
}

const (
	dockerHubAPI = "https://registry-1.docker.io/v2/"
)

type TokenResponse struct {
	Token string `json:"token"`
}

type Manifest struct {
	Config    Descriptor   `json:"config"`
	Layers    []Descriptor `json:"layers"`
	MediaType string       `json:"mediaType"`
}

type Descriptor struct {
	MediaType string `json:"mediaType"`
	Size      int64  `json:"size"`
	Digest    string `json:"digest"`
}

func getDockerHubToken(repo string) (string, error) {
	authURL := fmt.Sprintf("https://auth.docker.io/token?service=registry.docker.io&scope=repository:%s:pull", repo)
	resp, err := http.Get(authURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to authenticate: %s", resp.Status)
	}

	var tokenResponse TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return "", err
	}

	return tokenResponse.Token, nil
}

func getManifest(repo, tag, token string) (*Manifest, error) {
	client := &http.Client{}
	url := fmt.Sprintf("%s%s/manifests/%s", dockerHubAPI, repo, tag)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get manifest: %s", resp.Status)
	}

	var manifest Manifest
	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return nil, err
	}

	return &manifest, nil
}

func downloadLayer(repo, token, digest, outputPath string, size int64) error {
	client := &http.Client{}
	url := fmt.Sprintf("%s%s/blobs/%s", dockerHubAPI, repo, digest)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download layer: %s", resp.Status)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	progressWriter := &ProgressWriter{Writer: file, Total: size}
	_, err = io.Copy(progressWriter, resp.Body)
	return err
}

type ProgressWriter struct {
	Writer     io.Writer
	Total      int64
	Downloaded int64
}

func (pw *ProgressWriter) Write(p []byte) (int, error) {
	n, err := pw.Writer.Write(p)
	pw.Downloaded += int64(n)
	pw.printProgress()
	return n, err
}

func (pw *ProgressWriter) printProgress() {
	percent := float64(pw.Downloaded) / float64(pw.Total) * 100
	fmt.Printf("\rDownloading... %.2f%% complete", percent)
}

func loadRegexPatterns(filename string) (map[string]*regexp.Regexp, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var patterns map[string]string
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&patterns); err != nil {
		return nil, err
	}

	regexPatterns := make(map[string]*regexp.Regexp)
	for name, pattern := range patterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("failed to compile regex %s: %v", name, err)
		}
		regexPatterns[name] = re
	}

	return regexPatterns, nil
}

func checkPatterns(content string, patterns map[string]*regexp.Regexp) map[string][]string {
	matches := make(map[string][]string)
	for name, re := range patterns {
		foundMatches := re.FindAllString(content, -1)
		if foundMatches != nil {
			matches[name] = foundMatches
		}
	}
	return matches
}

func extractTarGz(tarGzPath, outputDir string) error {
	file, err := os.Open(tarGzPath)
	if err != nil {
		return err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tarReader := tar.NewReader(gzr)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(outputDir, header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.ModePerm); err != nil {
				return err
			}
		case tar.TypeReg:
			outFile, err := os.Create(target)
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		default:
			//fmt.Printf("Unable to untar type: %c in file %s", header.Typeflag, header.Name)
		}
	}
	return nil
}

func loadIgnoreExtensions(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var ignoreExtensions IgnoreExtensions
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&ignoreExtensions); err != nil {
		return nil, err
	}

	return ignoreExtensions.Extensions, nil
}

func shouldSkipFile(filename string, ignoreExtensions []string) bool {
	for _, ext := range ignoreExtensions {
		if strings.HasSuffix(strings.ToLower(filename), ext) {
			return true
		}
	}
	return false
}

func removeDir(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil
	}
	return os.RemoveAll(dir)
}

func printBanner() {
	banner := `
╭━━━━━━━━╮┏━╮╭━┓
┃┈┈┈┈┈┈┈┈┃╰╮╰╯╭╯   v1.1
┃╰╯┈┈┈┈┈┈╰╮╰╮╭╯┈   DOCKERSPY by Alisson Moretto (UndeadSec)
┣━━╯┈┈┈┈┈┈╰━╯┃┈┈         AUTOMATED OSINT ON DOCKER HUB     
╰━━━━━━━━━━━━╯┈┈`
	fmt.Println(color.New(color.FgGreen).Sprint(banner))
}

func fetchPaginatedResults(url string) ([]struct {
	Name        string `json:"repo_name"`
	Description string `json:"short_description"`
	PullCount   int    `json:"pull_count"`
	StarCount   int    `json:"star_count"`
	IsOfficial  bool   `json:"is_official"`
}, error) {
	var allResults []struct {
		Name        string `json:"repo_name"`
		Description string `json:"short_description"`
		PullCount   int    `json:"pull_count"`
		StarCount   int    `json:"star_count"`
		IsOfficial  bool   `json:"is_official"`
	}

	count := 0
	for {
		if count >= 100 {
			break
		}

		resp, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("API response error: %s", resp.Status)
		}

		var searchResult SearchResult
		if err := json.NewDecoder(resp.Body).Decode(&searchResult); err != nil {
			return nil, err
		}

		allResults = append(allResults, searchResult.Results...)
		count += len(searchResult.Results)

		if searchResult.Next == "" {
			break
		}

		url = searchResult.Next
	}

	if len(allResults) > 100 {
		allResults = allResults[:100]
	}

	return allResults, nil
}

func main() {
	printBanner()

	err := removeDir("docker_image")
	if err != nil {
		fmt.Println("\nError removing docker_image directory:", err)
		return
	}

	regexPatterns, err := loadRegexPatterns("/etc/dockerspy/configs/regex_patterns.json")
	if err != nil {
		fmt.Println("\nError loading regex patterns:", err)
		return
	}

	ignoreExtensions, err := loadIgnoreExtensions("/etc/dockerspy/configs/ignore_extensions.json")
	if err != nil {
		fmt.Println("\nError loading ignore extensions:", err)
		return
	}

	scanner := bufio.NewScanner(os.Stdin)
	info := color.New(color.FgCyan).SprintFunc()
	warning := color.New(color.FgYellow).SprintFunc()
	errorColor := color.New(color.FgRed).SprintFunc()
	success := color.New(color.FgGreen).SprintFunc()
	highlight := color.New(color.FgHiMagenta, color.Bold).SprintFunc()

	for {
		fmt.Print(info("\nEnter search term (or 'exit' to quit): "))
		scanner.Scan()
		searchTerm := scanner.Text()

		if strings.ToLower(searchTerm) == "exit" {
			break
		}

		dockerHubURL := "https://hub.docker.com/v2/search/repositories"
		params := url.Values{}
		params.Add("query", searchTerm)

		searchURL := fmt.Sprintf("%s?%s", dockerHubURL, params.Encode())
		results, err := fetchPaginatedResults(searchURL)
		if err != nil {
			fmt.Println(errorColor("\nError fetching search results:"), err)
			continue
		}

		fmt.Printf(info("\nFound %d results for '%s':"), len(results), searchTerm)
		for i, result := range results {
			fmt.Printf("\n%s - Name: %s\nDescription: %s\nStars: %d\nOfficial: %t", highlight(i+1), result.Name, result.Description, result.StarCount, result.IsOfficial)
		}

		fmt.Print(info("\nChoose a number or enter the full name to view repository tags (or 'cancel' to search again): "))
		scanner.Scan()
		choice := scanner.Text()

		if strings.ToLower(choice) == "cancel" {
			continue
		}

		var selectedRepo string
		choiceNum, err := strconv.Atoi(choice)
		if err == nil && choiceNum >= 1 && choiceNum <= len(results) {
			selectedRepo = results[choiceNum-1].Name
		} else {
			selectedRepo = choice
		}

		tagsURL := fmt.Sprintf("https://hub.docker.com/v2/repositories/%s/tags", selectedRepo)
		resp, err := http.Get(tagsURL)
		if err != nil {
			fmt.Println(errorColor("\nError fetching tags:"), err)
			continue
		}
		defer resp.Body.Close()

		var tagsResult TagsResult
		if err := json.NewDecoder(resp.Body).Decode(&tagsResult); err != nil {
			fmt.Println(errorColor("\nError decoding JSON response:"), err)
			continue
		}

		fmt.Printf(info("Available tags for repository '%s':"), selectedRepo)
		for i, tag := range tagsResult.Results {
			fmt.Printf("\n%s - %s", highlight(i+1), tag.Name)
		}

		fmt.Print(info("\nChoose a number to download the tag (or 'cancel' to search again): "))
		scanner.Scan()
		tagChoice := scanner.Text()

		if strings.ToLower(tagChoice) == "cancel" {
			continue
		}

		tagChoiceNum, err := strconv.Atoi(tagChoice)
		if err != nil || tagChoiceNum < 1 || tagChoiceNum > len(tagsResult.Results) {
			fmt.Println(warning("\nInvalid choice. Please try again."))
			continue
		}

		tag := tagsResult.Results[tagChoiceNum-1].Name

		repo := selectedRepo
		outputDir := "./docker_image"

		token, err := getDockerHubToken(repo)
		if err != nil {
			fmt.Println("\nError getting token:", err)
			return
		}

		manifest, err := getManifest(repo, tag, token)
		if err != nil {
			fmt.Println("\nError getting manifest:", err)
			return
		}

		os.MkdirAll(outputDir, os.ModePerm)

		var envContent string
		matchesResult := make(map[string]map[string][]string)

		for _, layer := range manifest.Layers {
			digestParts := strings.Split(layer.Digest, ":")
			if len(digestParts) != 2 {
				fmt.Println("\nInvalid digest format:", layer.Digest)
				continue
			}
			outputPath := filepath.Join(outputDir, digestParts[1]+".tar.gz")
			fmt.Println("\nDownloading layer:", layer.Digest)
			if err := downloadLayer(repo, token, layer.Digest, outputPath, layer.Size); err != nil {
				fmt.Println("\nError downloading layer:", err)
				return
			}

			extractedDir := filepath.Join(outputDir, digestParts[1])
			fmt.Println("\nExtracting layer:", outputPath)
			if err := extractTarGz(outputPath, extractedDir); err != nil {
				fmt.Println("\nError extracting layer:", err)
				continue
			}

			filepath.Walk(extractedDir, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if !info.IsDir() && !shouldSkipFile(path, ignoreExtensions) {
					content, err := os.ReadFile(path)
					if err != nil {
						fmt.Println("\nError reading file:", err)
						return nil
					}
					if filepath.Base(path) == ".env" {
						fmt.Println(success("\nFound .env file:"))
						envContent = string(content)
						fmt.Println(envContent)
					}
					matches := checkPatterns(string(content), regexPatterns)
					if len(matches) > 0 {
						fmt.Println(success("\nMatches found in file:"), path)
						matchesResult[path] = matches
						for pattern, matchedStrings := range matches {
							fmt.Printf("  Pattern: %s\n", pattern)
							for _, match := range matchedStrings {
								fmt.Printf("    %s\n", match)
							}
						}
					}
				}
				return nil
			})
		}

		fmt.Println(success("\nImage downloaded and extracted successfully\n"))

		resultData := map[string]interface{}{
			"selectedRepo": selectedRepo,
			"selectedTag":  tag,
			"envContent":   envContent,
			"matches":      matchesResult,
		}

		jsonFile, err := os.Create("results.json")
		if err != nil {
			fmt.Println(errorColor("\nError creating JSON file:"), err)
			return
		}
		defer jsonFile.Close()

		encoder := json.NewEncoder(jsonFile)
		if err := encoder.Encode(resultData); err != nil {
			fmt.Println(errorColor("\nError encoding JSON:"), err)
			return
		}

		fmt.Println(success("Results saved to results.json"))
	}
}
