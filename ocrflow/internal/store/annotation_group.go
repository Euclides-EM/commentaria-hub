package store

import (
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/annotation"
)

type AnnotationGroupSql struct {
	BaseSQL
}

func (s *AnnotationGroupSql) List() ([]*annotation.Group, error) {

}

func (s *AnnotationGroupSql) Get(id string) (*annotation.Group, error) {

}

func (s *AnnotationGroupSql) Create(group *annotation.Group) (*annotation.Group, error) {

}

func (s *AnnotationGroupSql) Update(group *annotation.Group) (*annotation.Group, error) {

}

func (s *AnnotationGroupSql) Delete(id string) error {

}
