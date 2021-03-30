package extensions

import (
	"fmt"
	"github.com/dghubble/sling"
	"net/http"
	"time"
)

// ExtDBReseeder reseeds a
type ExtDBReseeder struct {
	client *sling.Sling
	number int
}

func NewExtDBReseeder(baseURL string, number int) *ExtDBReseeder {
	// The HTTP client does not automatically set a timeout, hence we
	// arbitrarily choose a timeout of ten seconds.
	// See: https://medium.com/@nate510/don-t-use-go-s-default-http-client-4804cb19f779
	client := &http.Client{
		Timeout: time.Minute * 2,
	}

	return &ExtDBReseeder{
		client: sling.New().Client(client).Base(baseURL),
		number: number,
	}
}

func (r *ExtDBReseeder) Reseed() {
	if _, err := r.client.Delete("/db").ReceiveSuccess(nil); err != nil {
		panic(fmt.Errorf("reseed encountered unexpected error on DELETE /db: %w", err))
	}

	if _, err := r.client.Put(fmt.Sprintf("/db/%d", r.number)).ReceiveSuccess(nil); err != nil {
		panic(fmt.Errorf("reseed encountered unexpected error on PUT /db/%d: %w", r.number, err))
	}
}
