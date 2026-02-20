import { useLocalStorage } from "usehooks-ts";
import { ActiveTab } from "./types";
import { DefinitionsTab } from "./DefinitionsTab";
import { ExecutionsTab } from "./ExecutionsTab";
import { PageContainer, PageHeader, Title, TabBar, TabButton } from "./styles";

export function FeaturesPage() {
  const [activeTab, setActiveTab] = useLocalStorage<ActiveTab>(
    "features-active-tab",
    "definitions",
  );

  return (
    <PageContainer>
      <PageHeader>
        <Title>Title Page Features</Title>
      </PageHeader>

      <TabBar>
        <TabButton
          active={activeTab === "definitions"}
          onClick={() => setActiveTab("definitions")}
        >
          Definitions
        </TabButton>
        <TabButton
          active={activeTab === "executions"}
          onClick={() => setActiveTab("executions")}
        >
          Executions
        </TabButton>
      </TabBar>

      {activeTab === "definitions" && <DefinitionsTab />}
      {activeTab === "executions" && <ExecutionsTab />}
    </PageContainer>
  );
}
