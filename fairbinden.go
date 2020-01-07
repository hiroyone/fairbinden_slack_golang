package fairbinden

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly"
	"github.com/parnurzeal/gorequest"
)

var (
	Trace   *log.Logger
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
)

type Field struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

type Action struct {
	Type  string `json:"type"`
	Text  string `json:"text"`
	Url   string `json:"url"`
	Style string `json:"style"`
}

type Attachment struct {
	Fallback     string   `json:"fallback"`
	Color        string   `json:"color"`
	PreText      string   `json:"pretext"`
	AuthorName   string   `json:"author_name"`
	AuthorLink   string   `json:"author_link"`
	AuthorIcon   string   `json:"author_icon"`
	Title        string   `json:"title"`
	TitleLink    string   `json:"title_link"`
	Text         string   `json:"text"`
	ImageURL     string   `json:"image_url"`
	Fields       []Field  `json:"fields"`
	Footer       string   `json:"footer"`
	FooterIcon   string   `json:"footer_icon"`
	Timestamp    int64    `json:"ts"`
	MarkdownIn   []string `json:"mrkdwn_in"`
	Actions      []Action `json:"actions"`
	CallbackID   string   `json:"callback_id"`
	ThumbnailURL string   `json:"thumb_url"`
}

type Payload struct {
	Parse       string       `json:"parse,omitempty"`
	Username    string       `json:"username,omitempty"`
	IconUrl     string       `json:"icon_url,omitempty"`
	IconEmoji   string       `json:"icon_emoji,omitempty"`
	Channel     string       `json:"channel,omitempty"`
	Text        string       `json:"text,omitempty"`
	LinkNames   string       `json:"link_names,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
	UnfurlLinks bool         `json:"unfurl_links,omitempty"`
	UnfurlMedia bool         `json:"unfurl_media,omitempty"`
	Markdown    bool         `json:"mrkdwn,omitempty"`
}

func Init(
	traceHandle io.Writer,
	infoHandle io.Writer,
	warningHandle io.Writer,
	errorHandle io.Writer) {

	Trace = log.New(traceHandle,
		"TRACE: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Info = log.New(infoHandle,
		"INFO: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Warning = log.New(warningHandle,
		"WARNING: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Error = log.New(errorHandle,
		"ERROR: ",
		log.Ldate|log.Ltime|log.Lshortfile)
}

/*
 This function returns Japanese Standard Time
*/
func nowToday() time.Time {
	// logger.info("now_today fn")
	jst, _ := time.LoadLocation("Asia/Tokyo")
	// Info.Println("The time zone is: ", jst.String())
	now := time.Now().In(jst)
	Info.Println("The current time is: ", now.String())
	if now.Hour() < 11 {
		now = now.AddDate(0, 0, -1)
	}
	return now
}

/*
 This function returns if it is a weekday today
*/
func checkWeekday(now time.Time) bool {
	x := false
	if now.Weekday() >= 1 && now.Weekday() <= 5 {
		x = true
		Info.Println("It is a Weekday today")
	} else {
		Info.Println("It is the weekend")
	}
	return x
}

/*
Get a today's url for fairbinden lunch
*/
func getDailyURL(now time.Time) (*string, error) {
	if checkWeekday(now) {
		year := now.Year()
		month := int(now.Month())
		day := now.Day()
		// Convert each value into string and join them into url path
		datePath := path.Join(strconv.Itoa(year), fmt.Sprintf("%02d", month), fmt.Sprintf("%02d", day))
		domain := "http://xn--jvrr89ebqs6yg.tokyo/"
		url, _ := url.Parse(domain)
		url.Path = path.Join(url.Path, datePath)
		dayURL := url.String()
		Info.Println("Today's URL is", dayURL)
		return &dayURL, nil
	}
	return nil, errors.New("no lunch on the weekend")
}

/*
Get a today's menu URL for fairbinden
ref: https://benjamincongdon.me/blog/2018/03/01/Scraping-the-Web-in-Golang-with-Colly-and-Goquery/
*/
func getDailyMenuURL(dayURL string) (*string, error) {
	var dayMenuURL string
	url, _ := url.Parse(dayURL)
	domain := url.Host
	fmt.Println("Allowed Domain:", domain)

	c := colly.NewCollector(
		// Restrict crawling to specific domains
		colly.AllowedDomains(domain),
		// Allow visiting the same page multiple times
		colly.AllowURLRevisit(),
		// Allow crawling to be done in parallel / async
		colly.Async(false),
	)
	c.Limit(&colly.LimitRule{
		// Filter domains affected by this rule
		DomainGlob: domain + "/*",
		// Set a delay between requests to these domains
		Delay: 1 * time.Second,
		// Add an additional random delay
		RandomDelay: 1 * time.Second,
	})

	c.OnHTML("h3.title", func(e *colly.HTMLElement) {
		// Extract the link from the anchor HTML element
		link := e.ChildAttr("a", "href")
		// Info.Println(link)
		// Tell the collector to visit the link
		if strings.Contains(link, dayURL) {
			// Info.Println("Found daily URL link: ", e.Request.AbsoluteURL(link))
			// Info.Println("Found text: ", e.Text)
			dayMenuURL = link
		}
	})
	c.Visit(dayURL)
	Info.Println("Daily Menu URL is ", dayMenuURL)
	if dayMenuURL != "" {
		return &dayMenuURL, nil
	}
	return nil, errors.New("No lunch menu URL was found")
}

/*
Get a today's title for menu
*/
func getTitle(dayMenuURL string) *string {
	url, _ := url.Parse(dayMenuURL)
	domain := url.Host
	Info.Println("Allowed Domain:", domain)
	var title string

	c := colly.NewCollector(
		// Restrict crawling to specific domains
		colly.AllowedDomains(domain),
		// Allow visiting the same page multiple times
		colly.AllowURLRevisit(),
		// Allow crawling to be done in parallel / async
		colly.Async(false),
	)
	c.Limit(&colly.LimitRule{
		// Filter domains affected by this rule
		DomainGlob: domain + "/*",
		// Set a delay between requests to these domains
		Delay: 1 * time.Second,
		// Add an additional random delay
		RandomDelay: 1 * time.Second,
	})

	c.OnHTML(".post_title", func(e *colly.HTMLElement) {
		title = e.Text
	})
	c.Visit(dayMenuURL)

	Info.Println("Title is: ", title)
	return &title
}

/*
Get a today's texts for menu
*/
func getMainTexts(dayMenuURL string) *string {
	url, _ := url.Parse(dayMenuURL)
	domain := url.Host
	Info.Println("Allowed Domain:", domain)
	var texts string

	c := colly.NewCollector(
		// Restrict crawling to specific domains
		colly.AllowedDomains(domain),
		// Allow visiting the same page multiple times
		colly.AllowURLRevisit(),
		// Allow crawling to be done in parallel / async
		colly.Async(false),
	)
	c.Limit(&colly.LimitRule{
		// Filter domains affected by this rule
		DomainGlob: domain + "/*",
		// Set a delay between requests to these domains
		Delay: 1 * time.Second,
		// Add an additional random delay
		RandomDelay: 1 * time.Second,
	})
	c.OnHTML(".post_content", func(e *colly.HTMLElement) {
		e.ForEach("p", func(_ int, elem *colly.HTMLElement) {
			// Add * before \n for the left-hand emphasis sign
			if strings.Contains(elem.Text, "\n") {
				astaTexts := strings.Replace(elem.Text, "\n", "*\n", -1)
				texts += "*" + astaTexts + " \n"
			} else {
				texts += elem.Text + " \n"
			}
		})
	})
	c.Visit(dayMenuURL)
	Info.Println("The texts: ", texts)
	return &texts
}

/*
Get a today's image for menu
*/
func getImageURL(dayMenuURL string) *string {
	url, _ := url.Parse(dayMenuURL)
	domain := url.Host
	fmt.Println("Allowed Domain:", domain)
	var imageURL string

	c := colly.NewCollector(
		// Restrict crawling to specific domains
		colly.AllowedDomains(domain),
		// Allow visiting the same page multiple times
		colly.AllowURLRevisit(),
		// Allow crawling to be done in parallel / async
		colly.Async(false),
	)
	c.Limit(&colly.LimitRule{
		// Filter domains affected by this rule
		DomainGlob: domain + "/*",
		// Set a delay between requests to these domains
		Delay: 1 * time.Second,
		// Add an additional random delay
		RandomDelay: 1 * time.Second,
	})

	c.OnHTML(".post_image", func(e *colly.HTMLElement) {
		// Extract the link from the anchor HTML element
		imageURL = e.ChildAttr("img", "src")
	})
	c.Visit(dayMenuURL)

	fmt.Println("The image is: ", imageURL)
	return &imageURL
}

/*
Get a today's date in Japanese
*/
func getJapaneseDate(now time.Time) string {
	yobiArray := [7]string{"日", "月", "火", "水", "木", "金", "土"}
	yobi := yobiArray[now.Weekday()] // => 木
	// date = '{}月{}日({})'.format(now.month,now.day, yobi)
	// logger.info("Today's date is {}".format(date))
	date := fmt.Sprintf("%d年%d月%d日(%s)", now.Year(), now.Month(), now.Day(), yobi)
	fmt.Println(date)
	return date
}

func redirectPolicyFunc(req gorequest.Request, via []gorequest.Request) error {
	return fmt.Errorf("Incorrect token (redirection)")
}

func send(webhookUrl string, proxy string, payload Payload) []error {
	request := gorequest.New().Proxy(proxy)
	resp, _, err := request.
		Post(webhookUrl).
		RedirectPolicy(redirectPolicyFunc).
		Send(payload).
		End()

	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return []error{fmt.Errorf("Error sending msg. Status: %v", resp.Status)}
	}

	return nil
}

/*
SendSlack Sends a scraped message to Slack
*/
func SendSlack(w http.ResponseWriter, r *http.Request) {
	Init(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)
	// Logging Examples
	// Trace.Println("I have something standard to say")
	// Info.Println("Special Information")
	// Warning.Println("There is something you need to know about")
	// Error.Println("Something has failed")

	Info.Println("Run SendSlack Function")
	env := os.Getenv("ENV")

	now := nowToday()
	// For debugging: weekday menu
	// now := time.Date(2019, 6, 27, 23, 59, 59, 0, time.UTC)

	var webhookURL string
	if env == "PRD" {
		webhookURL = os.Getenv("channelPRD")
	} else if env == "STG" {
		webhookURL = os.Getenv("channelSTG")
	} else {
		Error.Println("The value must be either PRD or STG")
	}

	if dayURL, err := getDailyURL(now); err != nil {
		// panic(err)
		Info.Println("No posting to Slack:", err)
		// Write a text to HTTP page
		w.Write([]byte(fmt.Sprint("No posting to Slack:", err)))
	} else {
		Info.Println("Get article data")
		// Daily Menu if exists
		if dayMenuURL, err := getDailyMenuURL(*dayURL); err != nil {
			// panic(err)
			Info.Println("No posting to Slack:", err)
			// Write a text to HTTP page
			w.Write([]byte(fmt.Sprint("No posting to Slack:", err)))
		} else {
			mainText := *getMainTexts(*dayMenuURL)
			title := *getTitle(*dayMenuURL)
			imageURL := *getImageURL(*dayMenuURL)

			Info.Println("Other meta data")
			japaneseDate := getJapaneseDate(now)
			unixTime := now.Unix()

			Info.Println("Today's lunch menu URL:", *dayMenuURL)
			lunchAction := Action{
				Type:  "button",
				Text:  "今日のランチ🍚",
				Url:   *dayMenuURL,
				Style: "primary",
			}

			var officeLunchAction Action
			officeLunchURL := os.Getenv("channelOfficeBen")
			Info.Println("Office Lunch URL: ", officeLunchURL)

			// OfficeLunch is not available on Friday in my company
			if now.Weekday() <= 4 {
				officeLunchAction = Action{
					Type:  "button",
					Text:  "やっぱり会社の弁当🍱",
					Url:   officeLunchURL,
					Style: "danger",
				}
			} else {
				officeLunchAction = Action{}
			}

			Info.Println("Prepare attachments for slack posting")
			attachments := Attachment{
				Fallback:   "Required plain-text summary of the attachment.",
				Color:      "#36a64f",
				PreText:    japaneseDate + "のランチです！",
				Actions:    []Action{lunchAction, officeLunchAction},
				AuthorName: "フェアビンデン GO!",
				AuthorLink: "http://xn--jvrr89ebqs6yg.tokyo/",
				AuthorIcon: "http://flickr.com/icons/bobby.jpg",
				Title:      title,
				TitleLink:  *dayMenuURL,
				Text:       mainText,
				ImageURL:   imageURL,
				// ThumbnailURL: "http://example.com/path/to/thumb.png",
				Footer: "税込800円 11:00-14:00",
				// FooterIcon: "https://platform.slack-edge.com/img/default_application_icon.png",
				Timestamp: unixTime,
			}

			Info.Println("Post a message to Slack")
			payload := Payload{
				Attachments: []Attachment{attachments},
			}
			err := send(webhookURL, "", payload)
			if len(err) > 0 {
				fmt.Printf("error: %s\n", err)
			}

			// Write a text to HTTP page
			w.Write([]byte((mainText)))
		}
	}
}
