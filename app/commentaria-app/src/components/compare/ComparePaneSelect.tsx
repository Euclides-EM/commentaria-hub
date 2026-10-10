import { MultiSelectDropdown } from '../core/MultiSelectDropdown.tsx'
import { COMPARE_PANES, type ComparePane } from '../../context/compareSides.ts'

interface ComparePaneSelectProps {
  selectedPanes: ComparePane[] | null
  setSelectedPanes: (panes: ComparePane[] | null) => void
}

export function ComparePaneSelect({
  selectedPanes,
  setSelectedPanes,
}: ComparePaneSelectProps) {
  return (
    <div className="flex items-center text-sm gap-x-2 flex-nowrap shrink-0">
      <div className="h-3 w-3 rotate-[-45deg] border-b border-r border-slate-600" />
      <div className="text-shadow-gray-800 font-semibold px-1">Display</div>
      <MultiSelectDropdown<ComparePane>
        allItems={COMPARE_PANES}
        selectedItems={selectedPanes}
        setSelectedItems={setSelectedPanes}
        itemsLabel="parts"
        showBulkActions={false}
        getItemLabel={(item) => (item === 'scan' ? 'Facsimile' : 'Contents')}
        getPickerLabel={({ selectedItems }) => {
          if (
            selectedItems == null ||
            selectedItems.length === COMPARE_PANES.length
          ) {
            return 'Facsimile and contents'
          }
          if (selectedItems.length === 0) return 'None'
          return selectedItems[0] === 'scan' ? 'Facsimile' : 'Contents'
        }}
      />
    </div>
  )
}
