import type { feature_Feature, feature_Revision } from '@hub-api'
import { HexColorPicker } from 'react-colorful'
import { Button } from '../core/Button.tsx'

export type FeatureEditState = {
  name: string
  description: string
  color: string
}

const formatDate = (value?: string) => {
  if (!value) return 'Unknown'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString()
}

const normalizeHexInput = (value: string) => {
  const cleaned = value.trim().replace(/^#/, '')
  if (!cleaned) return ''
  return `#${cleaned.toLowerCase()}`
}

function ColorSwatch({ color }: { color?: string }) {
  return (
    <span
      className="w-3.5 h-3.5 rounded shrink-0 border border-gray-300"
      style={{
        backgroundColor: color || '#f2f2f2',
        boxShadow: 'inset 0 0 0 1px rgba(255,255,255,0.4)',
      }}
    />
  )
}

interface FeatureCardProps {
  feature: feature_Feature
  edits: FeatureEditState | undefined
  isEditing: boolean
  isExpanded: boolean
  isSaving: boolean
  isDirty: boolean
  isAuthenticated: boolean
  onSubmit: (event: React.FormEvent<HTMLFormElement>) => void
  onEdit: () => void
  onCancelEdit: () => void
  onDelete: () => void
  onToggleExpand: () => void
  onEditField: (update: Partial<FeatureEditState>) => void
  onNewRevision: (latestRevision?: feature_Revision) => void
}

export function FeatureCard({
  feature,
  edits,
  isEditing,
  isExpanded,
  isSaving,
  isDirty,
  isAuthenticated,
  onSubmit,
  onEdit,
  onCancelEdit,
  onDelete,
  onToggleExpand,
  onEditField,
  onNewRevision,
}: FeatureCardProps) {
  const revisions = feature.revisions ?? []
  const sortedRevisions = [...revisions].sort((a, b) => {
    const timeA = a.created_at ? new Date(a.created_at).getTime() : 0
    const timeB = b.created_at ? new Date(b.created_at).getTime() : 0
    return timeB - timeA
  })
  const latestRevision = sortedRevisions[0]

  return (
    <div className="bg-white border border-gray-200 rounded-xl p-5 shadow-sm flex flex-col gap-2">
      <form onSubmit={onSubmit}>
        <div className="flex justify-between items-start gap-4">
          <div className="flex items-center gap-2.5">
            <ColorSwatch color={feature.color} />
            <div className="text-base font-bold">
              {feature.name || 'Untitled'}
            </div>
          </div>
          <div className="flex items-center gap-2 shrink-0">
            {feature.is_default && (
              <span className="bg-teal-100 text-teal-700 rounded-full px-3 py-1 text-xs font-semibold">
                Default
              </span>
            )}
            {isAuthenticated && (
              <>
                {isEditing ? (
                  <>
                    <Button
                      type="submit"
                      variant="primary"
                      className="px-3 py-1.5 text-xs"
                      disabled={isSaving || !isDirty}
                    >
                      {isSaving ? 'Saving...' : 'Save'}
                    </Button>
                    <Button
                      type="button"
                      className="px-3 py-1.5 text-xs"
                      onClick={onCancelEdit}
                      disabled={isSaving}
                    >
                      Cancel
                    </Button>
                  </>
                ) : (
                  <Button
                    type="button"
                    className="px-3 py-1.5 text-xs"
                    onClick={onEdit}
                  >
                    Edit
                  </Button>
                )}
                {!isEditing && (
                  <Button
                    variant="danger"
                    type="button"
                    className="px-3 py-1.5 text-xs"
                    onClick={onDelete}
                    disabled={isSaving}
                  >
                    Delete
                  </Button>
                )}
              </>
            )}
          </div>
        </div>
        {isEditing ? (
          <div className="flex flex-col gap-1 mt-2">
            <div className="flex flex-col gap-1">
              <label className="text-xs text-gray-500 font-semibold">
                Name
              </label>
              <input
                className="px-2.5 py-2 border border-gray-300 rounded bg-white text-sm focus:outline-none focus:border-teal-500"
                value={edits?.name || ''}
                onChange={(event) => onEditField({ name: event.target.value })}
                disabled={isSaving}
                required
              />
            </div>
            <div className="flex flex-col gap-1">
              <label className="text-xs text-gray-500 font-semibold">
                Description
              </label>
              <textarea
                className="px-2.5 py-2 border border-gray-300 rounded bg-white text-sm resize-y min-h-[80px] focus:outline-none focus:border-teal-500"
                value={edits?.description || ''}
                onChange={(event) =>
                  onEditField({ description: event.target.value })
                }
                disabled={isSaving}
              />
            </div>
            <div className="flex flex-col gap-1">
              <label className="text-xs text-gray-500 font-semibold">
                Color
              </label>
              <div className="flex items-center gap-3 flex-wrap">
                <ColorSwatch color={edits?.color || feature.color} />
                <HexColorPicker
                  color={edits?.color || '#f2f2f2'}
                  onChange={(value) => onEditField({ color: value })}
                  style={{ width: 220, height: 170 }}
                />
                <input
                  className="px-2.5 py-2 border border-gray-300 rounded bg-white text-sm w-[120px] font-mono focus:outline-none focus:border-teal-500"
                  value={(edits?.color || '').replace(/^#/, '')}
                  onChange={(event) =>
                    onEditField({
                      color: normalizeHexInput(event.target.value),
                    })
                  }
                  disabled={isSaving}
                  aria-label="Feature color value"
                />
              </div>
            </div>
          </div>
        ) : (
          <div className="text-sm text-gray-500 leading-relaxed whitespace-pre-wrap mt-1">
            {feature.description || 'No description.'}
          </div>
        )}
      </form>
      <div className="text-xs text-gray-500">
        Last updated: {formatDate(feature.updated_at)}
      </div>
      <button
        type="button"
        className="border-none bg-transparent p-0 flex items-center gap-0 -ml-2 cursor-pointer"
        onClick={onToggleExpand}
        aria-label={isExpanded ? 'Collapse revisions' : 'Expand revisions'}
      >
        <span className="text-teal-700 rounded-full w-6 h-6 inline-flex items-center justify-center text-base hover:bg-gray-100 transition-colors">
          {isExpanded ? '▾' : '▸'}
        </span>
        <span className="text-xs text-gray-500">
          Revisions ({revisions.length})
        </span>
      </button>

      {isExpanded && (
        <div className="flex flex-col gap-3 mt-1">
          {isAuthenticated && (
            <Button
              type="button"
              variant="primary"
              className="px-3 py-1.5 text-xs self-start"
              onClick={() => onNewRevision(latestRevision)}
            >
              New revision
            </Button>
          )}
          {sortedRevisions.length === 0 ? (
            <div className="text-sm text-gray-500">
              No revisions available yet.
            </div>
          ) : (
            sortedRevisions.map((revision, index) => (
              <div
                key={revision.id ?? revision.created_at}
                className="bg-gray-100 border border-gray-200 rounded-lg px-3.5 py-3 text-xs grid gap-1.5 relative whitespace-pre-wrap"
              >
                <div>
                  <span className="font-semibold">ID:</span>{' '}
                  {revision.id || '—'}
                </div>
                {index === 0 && (
                  <span className="absolute top-2.5 right-3 bg-teal-100 text-teal-700 rounded-full px-3 py-1 text-xs font-semibold">
                    Latest
                  </span>
                )}
                {revision.prompt && (
                  <div>
                    <span className="font-semibold">Prompt:</span>
                    {'\n'}
                    {revision.prompt || '—'}
                  </div>
                )}
                {revision.categorizer && (
                  <div>
                    <span className="font-semibold">Categorizer:</span>
                    {'\n'}
                    {revision.categorizer || '—'}
                  </div>
                )}
                <div className="text-xs text-gray-500">
                  Created: {formatDate(revision.created_at)}
                </div>
              </div>
            ))
          )}
        </div>
      )}
    </div>
  )
}
