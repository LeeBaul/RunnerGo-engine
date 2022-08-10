package client

import (
	"github.com/olivere/elastic"
)

func NewEsClient(host string) (err error, esClient *elastic.Client) {
	esClient, err = elastic.NewClient(
		elastic.SetSniff(false),
		elastic.SetURL(host),
	)
	if err != nil {
		return
	}
	return
}
