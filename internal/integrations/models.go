package integrations

import "time"

type Maintainer struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Website string `json:"website"`
}

type Integration struct {
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	Logo           string    `json:"logo"`
	Category       string    `json:"category"`
	Version        string    `json:"version"`
	Tags           []string  `json:"tags"`
	Yml            string    `json:"yml"`
	RepositoryPath string    `json:"repositoryPath"`
	CreatedAt      time.Time `json:"createdAt"`
	//Maintainer     *Maintainer `json:"maintainer"`
	Vendor *Maintainer `json:"vendor"`
}
