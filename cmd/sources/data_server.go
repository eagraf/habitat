package sources

import "net/url"

type DataServerNode struct {
	Host string
	Port string
}

func (n *DataServerNode) GetUrl() (*url.URL, error) {
	return url.Parse("http://" + n.Host + ":" + n.Port)
}
