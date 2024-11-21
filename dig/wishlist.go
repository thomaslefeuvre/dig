package dig

import (
	"fmt"

	"github.com/thomaslefeuvre/bandcamp-tools/bandcamp"
)

type Wishlist struct {
	User *bandcamp.User
}

func NewWishlist(user *bandcamp.User) *Wishlist {
	return &Wishlist{
		User: user,
	}
}

func (b *Wishlist) Collect(r *Result) (*Result, error) {
	if r == nil {
		return nil, fmt.Errorf("uninitialised result")
	}

	if r.Full() {
		return r, nil
	}

	urls, err := b.User.GetWishlist()
	if err != nil {
		return r, err
	}

	for _, url := range urls {
		err = r.Add(url)
		if err != nil {
			return r, err
		}
	}

	return r, nil
}
