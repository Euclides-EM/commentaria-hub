package escriptorium

import (
	"fmt"
	"time"
)

type Document struct {
	PK                int             `json:"pk,omitempty"`
	Name              string          `json:"name"`
	Project           string          `json:"project"`
	Transcriptions    []Transcription `json:"transcriptions,omitempty"`
	MainScript        *string         `json:"main_script"`    // null or "Latin"
	ReadDirection     string          `json:"read_direction"` // "ltr", etc
	LineOffset        int             `json:"line_offset"`
	ShowConfidenceViz bool            `json:"show_confidence_viz"`
	ValidBlockTypes   []PartType      `json:"valid_block_types,omitempty"`
	ValidLineTypes    []PartType      `json:"valid_line_types,omitempty"`
	ValidPartTypes    []PartType      `json:"valid_part_types,omitempty"`
	PartsCount        int             `json:"parts_count,omitempty"`
	Tags              []string        `json:"tags"` // [] in JSON -> assume []string
	CreatedAt         time.Time       `json:"created_at,omitempty"`
	UpdatedAt         time.Time       `json:"updated_at,omitempty"`
}

type Transcription struct {
	PK            int      `json:"pk"`
	Name          string   `json:"name"`
	Archived      bool     `json:"archived"`
	AvgConfidence *float64 `json:"avg_confidence"` // null or float
}

func (c *Client) GetDocuments() ([]*Document, error) {
	return get[*Document](c, fmt.Sprintf("api/documents"))
}

//func (c *Client) CreateDocument(d *Document) (*Document, error) {
//	return post[*Document](c, "api/documents?format=api", d)
//}
//
//func (c *Client) CreateDocumentWithDefaults(name, projectName string) (*Document, error) {
//	ps, err := c.GetProjects()
//	if err != nil {
//		return nil, err
//	}
//
//	var projectSlug string
//	for _, p := range ps {
//		if p.Name == projectName || p.Slug == projectName {
//			projectSlug = p.Slug
//			break
//		}
//	}
//	if projectSlug == "" {
//		return nil, fmt.Errorf("project with name or slug '%s' not found", projectName)
//	}
//
//	lat := "Latin"
//	return c.CreateDocument(&Document{
//		Name:              name,
//		Project:           projectSlug,
//		ReadDirection:     "ltr",
//		LineOffset:        0,
//		MainScript:        &lat,
//		ShowConfidenceViz: true,
//		Tags:              []string{},
//	})
//}
//
//func (c *Client) DeleteDocument(documentID int) error {
//	return doDelete(c, fmt.Sprintf("/api/documents/%d", documentID))
//}
//
//func (c *Client) DeleteDocumentByName(name string) error {
//	ds, err := c.GetDocuments()
//	if err != nil {
//		return err
//	}
//
//	for _, d := range ds {
//		if d.Name == name {
//			return c.DeleteDocument(d.PK)
//		}
//	}
//
//	return fmt.Errorf("document with name or project '%s' not found", name)
//}
