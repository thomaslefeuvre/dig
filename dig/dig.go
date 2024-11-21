package dig

import (
	"encoding/json"
	"fmt"
	neturl "net/url"
	"os"
	"os/exec"
	"strings"
	"time"
)

type Dig struct {
	Collectors []Collector
	Cap        int64
}

type Collector interface {
	Collect(*Result) (*Result, error)
}

type Result struct {
	urls map[string]struct{}
	cap  int64

	Albums []string `json:"albums,omitempty"`
	Tracks []string `json:"tracks,omitempty"`
}

func NewResult(cap int64) *Result {
	return &Result{
		urls: map[string]struct{}{},
		cap:  cap,
	}
}

func NewResultFromFile(inputFile string) (*Result, error) {
	f, err := os.Open(inputFile)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	var result Result
	dec := json.NewDecoder(f)
	if err := dec.Decode(&result); err != nil {
		return nil, fmt.Errorf("decode json: %w", err)
	}
	result.initialise()

	return &result, nil
}

func (r *Result) initialise() {
	r.urls = map[string]struct{}{}
	for _, url := range r.Albums {
		r.urls[url] = struct{}{}
	}
	for _, url := range r.Tracks {
		r.urls[url] = struct{}{}
	}
	r.cap = int64(len(r.urls))
}

func (r *Result) Add(url string) error {
	if r == nil {
		return fmt.Errorf("uninitialised result")
	}

	if r.Full() {
		return nil
	}

	u, err := neturl.Parse(url)
	if err != nil {
		return fmt.Errorf("parsing url: %w", err)
	}

	u.RawQuery = ""

	if _, ok := r.urls[u.String()]; ok {
		// already added
		return nil
	}

	if strings.HasPrefix(u.Path, "/track") {
		r.urls[u.String()] = struct{}{}
		r.Tracks = append(r.Tracks, u.String())
	}

	if strings.HasPrefix(u.Path, "/album") {
		r.urls[u.String()] = struct{}{}
		r.Albums = append(r.Albums, u.String())
	}

	return nil
}

func (r *Result) Save(dir string) (filename string, err error) {
	// Generate filename using struct type and timestamp
	filename = fmt.Sprintf("dig-%d.json", time.Now().Unix())

	// Open file for writing
	f, err := os.Create(fmt.Sprintf("%s/%s", dir, filename))
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(&r); err != nil {
		return "", fmt.Errorf("failed to encode struct to JSON: %w", err)
	}

	return filename, nil
}

func (r *Result) OpenInBrowser() error {
	args := []string{"-na", "Microsoft Edge", "--args", "--new-window"}
	args = append(args, r.urlSlice()...)

	cmd := exec.Command("open", args...)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("opening tabs: %w", err)
	}

	return nil
}

func (r *Result) urlSlice() []string {
	urls := []string{}
	for url := range r.urls {
		urls = append(urls, url)
	}
	return urls
}

func (r *Result) Full() bool {
	return r.cap > 0 && int64(len(r.urls)) >= r.cap
}

func New(cap int64, collectors ...Collector) *Dig {
	return &Dig{
		Collectors: collectors,
		Cap:        cap,
	}
}

func (d *Dig) Run() (*Result, error) {
	result := NewResult(d.Cap)
	var err error

	for _, c := range d.Collectors {
		result, err = c.Collect(result)
		if err != nil {
			// print the error here but continue
			fmt.Println(err)
			break
		}
	}

	return result, err
}

func ReleaseURL(url neturl.URL) bool {
	if strings.HasPrefix(url.Path, "/track") {
		return true
	}

	if strings.HasPrefix(url.Path, "/album") {
		return true
	}

	return false
}
