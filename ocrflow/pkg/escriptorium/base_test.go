package escriptorium

import (
	"fmt"
	"testing"
)

func TestUploadImage(t *testing.T) {
	pageToUpload := 158
	client := NewClient("admin", "admin", "http://localhost:8080/")
	err := client.UploadImage("Test1", fmt.Sprintf("/Users/mia/dev/personal/elements-dh/ocrflow/store/data/aiqcec/imgs/page-%04d.png", pageToUpload))
	if err != nil {
		t.Fatalf("CreateDocumentWithDefaults failed: %v", err)
	}
	fmt.Printf("Successfully uploaded page %d.\n", pageToUpload)
	err = client.UploadAnnotation("Test1", fmt.Sprintf("/Users/mia/dev/personal/elements-dh/ocrflow/pkg/escriptorium/testdata/3rtoer/alto/page-%04d.xml", pageToUpload))
	if err != nil {
		t.Fatalf("UploadAnnotation failed: %v", err)
	}
	fmt.Printf("Successfully uploaded annotation for page %d.\n", pageToUpload)
}
