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

	if err = g.verifyReferencedAnnotations(group.Annotations); err != nil {
		return nil, fmt.Errorf("failed to verify annotation references: %w", err)
	}

	group.ID = idgen.GenerateID("anng")
	group.Name = name.NextAvailable(existingGroupNames, group.Name)
	return g.annotationGroupStore.Create(group)
}

func (g *AnnotationGroup) verifyReferencedAnnotations(refsInRequest []*annotation.Reference) error {
	// check for duplicates in refsInRequest
	if len(lo.UniqBy(refsInRequest, func(r *annotation.Reference) string {
		return fmt.Sprintf("%s:%s", r.DatasetID, r.ID)
	})) != len(refsInRequest) {
		return errors.New("duplicate annotation references found in request")
	}
	refs, err := g.annotationSvc.ListAnnotationsByAnnotationReferences(refsInRequest)
	if err != nil {
		return err
	}
	if len(refs) != len(refsInRequest) {
		var nonExistentIDs []*annotation.Reference
		for _, id := range refsInRequest {
			if !lo.ContainsBy(refs, func(r *annotation.Annotation) bool {
				return r.ID == id.ID && r.DatasetID == id.DatasetID
			}) {
				nonExistentIDs = append(nonExistentIDs, id)
			}
		}
		return fmt.Errorf("some annotation references do not exist: %v", nonExistentIDs)
	}
	return nil
}

func (g *AnnotationGroup) Update(id string, group *annotation.Group) (*annotation.Group, error) {
	existingGroup, err := g.Get(id)
	if err != nil {
		return nil, err
	}

	if err := g.verifyReferencedAnnotations(group.Annotations); err != nil {
		return nil, fmt.Errorf("failed to verify annotation references: %w", err)
	}

	group.ID = id
	if group.Name != existingGroup.Name {
		existingGroupNames, err := g.getExistingGroupNames()
		if err != nil {
			return nil, err
		}
		group.Name = name.NextAvailable(existingGroupNames, group.Name)
	}
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
