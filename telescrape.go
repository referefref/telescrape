package main

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/fatih/color"
)

type Attachment struct {
	URL      string `json:"url"`
	Filename string `json:"filename"`
	FileHash string `json:"file_hash"`
}

type PostData struct {
	Author      string       `json:"author"`
	DateTime    string       `json:"date_time"`
	Content     string       `json:"content"`
	Attachments []Attachment `json:"attachments"`
	Links       []string     `json:"links"`
}

type TeleScraper struct {
	postURL  string
	postData PostData
	headers  map[string]string
}

func NewTeleScraper(url string) *TeleScraper {
	return &TeleScraper{
		postURL: url,
		headers: map[string]string{
			//"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/77.0.3865.90 Safari/537.36 TelegramBot (like TwitterBot)",
			"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36 OPR/107.0.0.0",
		},
	}
}

func (ts *TeleScraper) fetchAndProcessPost() {
	var mainData, additionalData struct {
		Title       string   `json:"title"`
		Image       string   `json:"image"`
		Description string   `json:"description"`
		Author      string   `json:"author"`
		Views       string   `json:"views"`
		Datetime    string   `json:"datetime"`
		Links       []string `json:"links"`
	}

	if err := ts.callScraper(ts.postURL, &mainData); err != nil {
		fmt.Println("Error scraping main content:", err)
		return
	}

	embedURL := ts.postURL + "?embed=1&mode=tme"
	if err := ts.callScraper(embedURL, &additionalData); err != nil {
		fmt.Println("Error scraping additional data:", err)
		return
	}

	ts.postData.Author = additionalData.Author
	ts.postData.Content = mainData.Description
	ts.postData.DateTime = additionalData.Datetime

	ts.postData.Attachments = append(ts.postData.Attachments, Attachment{
		URL: mainData.Image,
	})

	ts.postData.Links = filterAndProcessLinks(ts.postData.Links)
	ts.promptForMediaDownload()
	ts.savePostDetails()
	fmt.Println("Finished processing URL.")
}

func (ts *TeleScraper) promptForMediaDownload() {
    fmt.Println("Starting media download...")
    ts.downloadMedia()
    fmt.Println("Media download completed.")
}

func filterAndProcessLinks(links []string) []string {
    var processedLinks []string
    for _, link := range links {
        if link != "" && !strings.Contains(link, "telegram.org") && !strings.Contains(link, "tg://resolve") {
            processedLinks = append(processedLinks, link)
        }
    }
    return processedLinks
}

func (ts *TeleScraper) callScraper(url string, data interface{}) error {
	cmd := exec.Command("node", "./pupeteer_scraper/scraper.js", url, "./pupeteer_scraper/cookie.txt")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to execute script: %v", err)
	}

	if err := json.Unmarshal(output, data); err != nil {
		return fmt.Errorf("failed to parse JSON output: %v", err)
	}

	return nil
}

func (ts *TeleScraper) processHTML(htmlContent string) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		color.Red("Error loading HTML content: %v", err)
		return
	}

	ts.postData.Author = doc.Find(".tgme_widget_message_owner_name span").Text()
	ts.postData.DateTime = doc.Find(".datetime").Text()
	ts.postData.Content, _ = doc.Find(".tgme_widget_message_text").Html()
	ts.postData.Content = strings.TrimSpace(htmlToText(ts.postData.Content))

	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if exists {
			ts.postData.Links = append(ts.postData.Links, href)
		}
	})

	fmt.Println("Processed HTML content successfully.")
}

func (ts *TeleScraper) downloadMedia() {
	for i, attachment := range ts.postData.Attachments {
		hasher := sha256.New()
		hasher.Write([]byte(attachment.URL))
		hash := hex.EncodeToString(hasher.Sum(nil))
		timestamp := time.Now().Format("20060102150405")
		uniqueFilename := fmt.Sprintf("%s_%s", timestamp, hash[:8]) + ".jpg" 

		filePath, err := downloadFile(attachment.URL, uniqueFilename)
		if err != nil {
			fmt.Printf("Failed to download or save file: %v\n", err)
			continue
		}

		hash, err = calculateFileHash(filePath)
		if err != nil {
			fmt.Printf("Failed to calculate hash for %s: %v\n", filePath, err)
			continue
		}

		ts.postData.Attachments[i].Filename = uniqueFilename
		ts.postData.Attachments[i].FileHash = hash

		fmt.Printf("Downloaded and hashed %s successfully\n", uniqueFilename)
	}
}

func downloadFile(url, filename string) (string, error) {
	tempDir := "./temp"
	if err := os.MkdirAll(tempDir, os.ModePerm); err != nil {
		return "", fmt.Errorf("failed to create temp directory: %v", err)
	}

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to download file: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status: %s", resp.Status)
	}

	filePath := filepath.Join(tempDir, filename)
	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to write file: %v", err)
	}

	return filePath, nil
}

func calculateFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func (ts *TeleScraper) savePostDetails() {
	jsonData, err := json.MarshalIndent(ts.postData, "", "  ")
	if err != nil {
		color.Red("Failed to marshal post details: %v", err)
		return
	}

	filename := "post_details.json"
	err = ioutil.WriteFile(filename, jsonData, 0644)
	if err != nil {
		color.Red("Failed to write post details to file: %v", err)
		return
	}
	color.Green("Post details saved to %s", filename)
}

func htmlToText(html string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	text := re.ReplaceAllString(html, "")
	return strings.TrimSpace(text)
}

func getFilenameFromURL(url string) string {
    segments := strings.Split(url, "/")
	    filename := segments[len(segments)-1]
    return filename
}

func main() {
	var target string
	flag.StringVar(&target, "t", "", "Telegram post URL")
	flag.Parse()

	if target == "" {
		fmt.Println("No target URL provided. Exiting.")
		return
	}

	scraper := NewTeleScraper(target)
	scraper.fetchAndProcessPost()
}
