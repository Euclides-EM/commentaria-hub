package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model/annotation"
	"github.com/MiaMish/elements-dh/ocrflow/internal/store"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/idgen"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/name"
	"github.com/samber/lo"
)

type AnnotationGroup struct {
	annotationSvc        *Annotation
	annotationGroupStore *store.AnnotationGroupSql
}

func NewAnnotationGroupService(annotationSvc *Annotation, annotationGroupStore *store.AnnotationGroupSql) *AnnotationGroup {
	return &AnnotationGroup{
		annotationSvc:        annotationSvc,
		annotationGroupStore: annotationGroupStore,
	}
}

func (g *AnnotationGroup) List() ([]*annotation.Group, error) {
	return g.annotationGroupStore.List()
}

func (g *AnnotationGroup) Get(id string) (*annotation.Group, error) {
	gr, err := g.annotationGroupStore.Get(id)
	if err != nil {
		return nil, err
	}
	if gr == nil {
		return nil, errors.New("annotation group not found")
	}
	return gr, nil
}

func (g *AnnotationGroup) Create(group *annotation.Group) (*annotation.Group, error) {
	existingGroupNames, err := g.getExistingGroupNames()
	if err != nil {
		return nil, err
	}

	group.ID = idgen.GenerateID("anng")
	group.Name = name.NextAvailable(existingGroupNames, group.Name)
	return g.annotationGroupStore.Create(group)
}

func (g *AnnotationGroup) Update(id string, group *annotation.Group) (*annotation.Group, error) {
	if _, err := g.Get(id); err != nil {
		return nil, err
	}

	existingGroupNames, err := g.getExistingGroupNames()
	if err != nil {
		return nil, err
	}

	group.Name = name.NextAvailable(existingGroupNames, group.Name)
	group.UpdatedAt = time.Now()

	return g.annotationGroupStore.Update(group)
}

func (g *AnnotationGroup) Delete(id string) error {
	if _, err := g.Get(id); err != nil {
		return err
	}

	return g.annotationGroupStore.Delete(id)
}

func (g *AnnotationGroup) getExistingGroupNames() ([]string, error) {
	groups, err := g.List()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve existing annotation groups: %w", err)
	}
	return lo.Map(groups, func(gr *annotation.Group, _ int) string {
		return gr.Name
	}), nil
}
