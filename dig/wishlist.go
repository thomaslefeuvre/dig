package dig

import (
	"fmt"

	"github.com/thomaslefeuvre/digg/bandcamp"
)

type Wishlist struct {
	User *bandcamp.User
}

func NewWishlist(user *bandcamp.User) *Wishlist {
	return &Wishlist{
		User: user,
	}
}

func (b *Wishlist) Collect(c *Collection) (*Collection, error) {
	if c == nil {
		return nil, fmt.Errorf("uninitialised result")
	}

	urls, err := b.User.GetWishlist()
	if err != nil {
		return c, err
	}

	for _, url := range urls {
		err = c.Add(url)
		if err != nil {
			return c, err
		}
	}

	return c, nil
}
