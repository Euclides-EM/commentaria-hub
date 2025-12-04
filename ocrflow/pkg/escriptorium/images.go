package escriptorium

import "fmt"

func (c *Client) UploadImage(documentIdentifier string, filePath string) error {
	documentId, err := c.inferDocumentID(documentIdentifier)
	if err != nil {
		return err
	}
	return upload(c, fmt.Sprintf("/api/documents/%d/parts/", documentId), "image", filePath, nil)
}

func (c *Client) inferDocumentID(documentIdentifier string) (int, error) {
	docs, err := c.GetDocuments()
	if err != nil {
		return 0, fmt.Errorf("error getting documents: %v", err)
	}
	for _, d := range docs {
		if fmt.Sprintf("%d", d.PK) == documentIdentifier || d.Name == documentIdentifier {
			return d.PK, nil
		}
	}
	return 0, fmt.Errorf("document %s not found", documentIdentifier)
}
