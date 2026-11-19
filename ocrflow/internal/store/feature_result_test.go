package store

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
)

func TestListEditionFeatureValuesReturnsCurrentClassification(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })

	_, err = db.Exec(`
CREATE TABLE edition_feature_results (
  scope TEXT NOT NULL,
  edition_id TEXT NOT NULL,
  feature_id TEXT NOT NULL,
  PRIMARY KEY (scope, edition_id, feature_id)
);
CREATE TABLE edition_feature_result_values (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  scope TEXT NOT NULL,
  edition_id TEXT NOT NULL,
  feature_id TEXT NOT NULL,
  surface TEXT NOT NULL
);
INSERT INTO edition_feature_results VALUES ('editions', 'ed-1', 'm_classifier');
INSERT INTO edition_feature_results VALUES ('editions', 'ed-1', 'another_feature');
INSERT INTO edition_feature_result_values (scope, edition_id, feature_id, surface) VALUES
  ('editions', 'ed-1', 'm_classifier', 'Geometry::primary'),
  ('editions', 'ed-1', 'm_classifier', 'Optics::secondary'),
  ('editions', 'ed-1', 'another_feature', 'ignored');
`)
	require.NoError(t, err)

	resultStore := NewFeatureResultSQL(db)
	values, err := resultStore.ListEditionFeatureValues("m_classifier", []string{"ed-1", "ed-2"})
	require.NoError(t, err)
	require.Equal(t, []string{"Geometry::primary", "Optics::secondary"}, values["ed-1"])
	require.NotContains(t, values, "ed-2")

	// A feature rerun replaces its value rows; the next read must see the replacement.
	_, err = db.Exec(`
DELETE FROM edition_feature_result_values
WHERE scope = 'editions' AND edition_id = 'ed-1' AND feature_id = 'm_classifier';
INSERT INTO edition_feature_result_values (scope, edition_id, feature_id, surface)
VALUES ('editions', 'ed-1', 'm_classifier', 'Arithmetic::primary');
`)
	require.NoError(t, err)

	values, err = resultStore.ListEditionFeatureValues("m_classifier", []string{"ed-1"})
	require.NoError(t, err)
	require.Equal(t, []string{"Arithmetic::primary"}, values["ed-1"])

	matchingIDs, err := resultStore.ListEditionIDsByFeatureValues(
		"m_classifier",
		[]string{"arithmetic::PRIMARY", "Geometry::primary"},
	)
	require.NoError(t, err)
	require.Equal(t, map[string]struct{}{"ed-1": {}}, matchingIDs)

	matchingIDs, err = resultStore.ListEditionIDsByFeatureValues(
		"m_classifier",
		[]string{"Geometry::primary"},
	)
	require.NoError(t, err)
	require.Empty(t, matchingIDs)
}
