import { parseAsStringLiteral, useQueryState } from 'nuqs'
import { TabButton } from '../core/TabButton'
import { FeatureExecutionsBrowser } from './FeatureExecutionsBrowser'
import { FeatureResultsBrowser } from './FeatureResultsBrowser'
import { FeaturesDefinitionsTab } from './FeaturesDefinitionsTab'

type FeaturesPageTab = 'features' | 'results' | 'executions'

export function FeaturesTable() {
  const [activeTab, setActiveTab] = useQueryState(
    'featuresTab',
    parseAsStringLiteral<FeaturesPageTab>([
      'features',
      'results',
      'executions',
    ]).withDefault('features'),
  )

  return (
    <div className="h-full flex flex-col overflow-hidden">
      <div className="flex w-full gap-4 p-3 border-b border-gray-200 bg-white items-center justify-center">
        <TabButton
          onSelected={() => void setActiveTab('features')}
          title="Features"
          isActive={activeTab === 'features'}
        />
        <TabButton
          onSelected={() => void setActiveTab('results')}
          title="Feature Results"
          isActive={activeTab === 'results'}
        />
        <TabButton
          onSelected={() => void setActiveTab('executions')}
          title="Feature Executions"
          isActive={activeTab === 'executions'}
        />
      </div>
      <div className="flex-1 min-h-0 overflow-auto p-3">
        {activeTab === 'features' && <FeaturesDefinitionsTab />}
        {activeTab === 'results' && <FeatureResultsBrowser />}
        {activeTab === 'executions' && <FeatureExecutionsBrowser />}
      </div>
    </div>
  )
}
