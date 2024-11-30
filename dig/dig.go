package dig

import (
	"encoding/json"
	"fmt"
	"log"
	neturl "net/url"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"
)

type Dig struct {
	Collectors []Collector
}

type Collector interface {
	Collect(*Collection) (*Collection, error)
}

type Collection struct {
	Releases map[string]bool `json:"releases"`

	dir string
}

func NewCollection(dir string) *Collection {
	return &Collection{
		Releases: map[string]bool{},

		dir: dir,
	}
}

func LoadCollection(dir string) (*Collection, error) {
	return LoadCollectionFromFile(dir, "collection.json")
}

func LoadCollectionFromFile(dir, filename string) (*Collection, error) {
	f, err := os.Open(fmt.Sprintf("%s/%s", dir, filename))
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	var c Collection
	dec := json.NewDecoder(f)
	if err := dec.Decode(&c); err != nil {
		return nil, fmt.Errorf("decode json: %w", err)
	}

	c.dir = dir

	return &c, nil
}

func (c *Collection) Add(url string) error {
	if c == nil {
		return fmt.Errorf("uninitialised result")
	}

	u, err := neturl.Parse(url)
	if err != nil {
		return fmt.Errorf("parsing url: %w", err)
	}

	u.RawQuery = ""

	if _, ok := c.Releases[u.String()]; ok {
		// already added
		return nil
	}

	if strings.HasPrefix(u.Path, "/track") || strings.HasPrefix(u.Path, "/album") {
		r := u.String()
		c.Releases[r] = false
		log.Printf("added %s to collection", r)
	}

	return nil
}

func (c *Collection) Save() (filenames []string, err error) {
	filenames = []string{
		fmt.Sprintf("collection-%d.json", time.Now().Unix()),
		"collection.json",
	}

	for _, name := range filenames {
		saveCollection := func(name string) error {
			// Open file for writing
			f, err := os.Create(fmt.Sprintf("%s/%s", c.dir, name))
			if err != nil {
				return fmt.Errorf("failed to create file: %w", err)
			}
			defer f.Close()

			enc := json.NewEncoder(f)
			enc.SetIndent("", "  ")
			if err := enc.Encode(&c); err != nil {
				return fmt.Errorf("failed to encode struct to JSON: %w", err)
			}

			return nil
		}
		err := saveCollection(name)
		if err != nil {
			return nil, err
		}
	}

	return filenames, nil
}

func (c *Collection) OpenInBrowser(quantity int) error {
	urls := c.List(quantity)
	err := openURLs(urls)
	if err != nil {
		return err
	}
	c.markOpened(urls)
	return nil
}

func openURLs(urls []string) error {
	args := []string{"-na", "Microsoft Edge", "--args", "--new-window"}
	args = append(args, urls...)

	cmd := exec.Command("open", args...)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("opening tabs: %w", err)
	}

	return nil
}

// quantity >= 0
func (c *Collection) List(limit int) []string {
	items := []string{}
	for item := range c.Releases {
		items = append(items, item)
		if len(items) == limit {
			break
		}
	}
	sort.Strings(items)
	return items
}

func (c *Collection) markOpened(urls []string) {
	for _, url := range urls {
		c.Releases[url] = true
	}
}

func New(collectors ...Collector) *Dig {
	return &Dig{
		Collectors: collectors,
	}
}

func (d *Dig) UpdateCollection(collection *Collection) *Collection {
	var err error

	log.Printf("updating collection (initial size: %d)", collection.Size())

	for _, c := range d.Collectors {
		collection, err = c.Collect(collection)
		if err != nil {
			// print the error here but continue
			log.Printf("error during collect: %v", err)
			break
		}
	}

	return collection
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

func (c *Collection) Size() int {
	return len(c.Releases)
}

func (c *Collection) All() []string {
	return c.List(-1)
}

func (c *Collection) Filter(query string) []string {
	items := []string{}
	for url := range c.Releases {
		if strings.Contains(url, query) {
			items = append(items, url)
		}
	}
	sort.Strings(items)
	return items
}
