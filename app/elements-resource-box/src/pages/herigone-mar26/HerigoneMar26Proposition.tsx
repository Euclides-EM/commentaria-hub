import { useState, type WheelEvent } from "react";
import styled from "@emotion/styled";
import { PANE_COLOR, SEA_COLOR } from "../../utils/colors";
import { type VisibleState } from "./visibleState";
import { ConstructionDiagram } from "./diagrams/ConstructionDiagram";
import { PropositionDiagram } from "./diagrams/PropositionDiagram";

type Step = {
  groupIndex: number;
  stepIndex: number;
  groupLabel: string;
  label: string;
  text: string;
  visualStep: number;
};

type StepGroup = {
  label: string;
  steps: Array<{
    text: string;
    visualStep: number;
  }>;
};

type SectionViewMode = "step-by-step" | "sequential";

const stepGroups: StepGroup[] = [
  {
    label: "I. The Hypothesis",
    steps: [
      { text: "AB is ————", visualStep: 1 },
      { text: "AC & CB are parts of AB", visualStep: 2 },
    ],
  },
  {
    label: "II. The Proposition",
    steps: [
      {
        text: "Required to demonstrate:\n□ AB = □ AC + □ CB + 2 ▭ ACB",
        visualStep: 3,
      },
    ],
  },
  {
    label: "III. The Preparation (1634)",
    steps: [
      { text: "AD is □ on AB", visualStep: 4 },
      {
        text: "AD is □ on AB",
        visualStep: 5,
      },
      { text: "EB is ————", visualStep: 6 },
      {
        text: "CF = AE = BD",
        visualStep: 7,
      },
      { text: "HGI = AB = ED", visualStep: 8 },
    ],
  },
  {
    label: "III. The Demonstration (1634)",
    steps: [
      {
        text: "∠EAB, ∠AED, ∠EDB, ∠ABD are 90°",
        visualStep: 9,
      },
      {
        text: "∠EHG, ∠EFG, ∠HGF are 90°",
        visualStep: 10,
      },
      { text: "AE = AB", visualStep: 11 },
      {
        text: "∠AEB, ∠DEB are 45°",
        visualStep: 12,
      },
      {
        text: "∠HGE, ∠FGE are 45°",
        visualStep: 13,
      },
      {
        text: "HE = HG = EF = FG, HF is □ on HG",
        visualStep: 14,
      },
      {
        text: "CGIB is □ on CB",
        visualStep: 15,
      },
      {
        text: "▭ AG = ▭ ACB = ▭ GD",
        visualStep: 16,
      },
      {
        text: "□ AD = □ HF + □ CI + ▭ AG + ▭ GD",
        visualStep: 17,
      },
      {
        text: "□ AB = □ AC + □ CB + 2 ▭ ACB",
        visualStep: 18,
      },
    ],
  },
  {
    label: "III. The Preparation (1639)",
    steps: [
      { text: "AD is □ on AB", visualStep: 4 },
      {
        text: "AD is □ on AB",
        visualStep: 5,
      },
      { text: "EB is ————", visualStep: 6 },
      {
        text: "CF ∥ AE",
        visualStep: 7,
      },
      { text: "HGI ∥ AB", visualStep: 8 },
    ],
  },
  {
    label: "V. The Demonstration (1639)",
    steps: [
      {
        text: "AG, HF, CI, GD are ▱",
        visualStep: 19,
      },
      {
        text: "∠A, ∠AED, ∠D, ∠ABD are 90°",
        visualStep: 20,
      },
      { text: "AE = AB", visualStep: 21 },
      {
        text: "∠AEB = ∠ABE = ∠DEB = ∠DBE",
        visualStep: 22,
      },
      {
        text: "∠CGB = ∠AEB, ∠HGE = ∠ABE, ∠IGB = ∠BED",
        visualStep: 23,
      },
      {
        text: "∠CBG = ∠CGB = ∠HEG = ∠HGE",
        visualStep: 24,
      },
      {
        text: "BC = CG, GH = HE",
        visualStep: 25,
      },
      {
        text: "HE = GF, HG = EF, CB = GI, CG = BI",
        visualStep: 26,
      },
      {
        text: "GHFE is □ on HG, CBIG is □ on CB",
        visualStep: 27,
      },
      {
        text: "AG & GD are ▭ on ACB and AG = GD",
        visualStep: 28,
      },
      {
        text: "□ AD = □ HF + □ CI + ▭ AG + ▭ GD",
        visualStep: 29,
      },
      {
        text: "□ AB = □ AC + □ CB + 2 ▭ ACB",
        visualStep: 30,
      },
    ],
  },
];

const steps: Step[] = stepGroups.flatMap(
  ({ label, steps: groupedSteps }, groupIndex) =>
    groupedSteps.map(({ text, visualStep }, index) => ({
      groupIndex,
      stepIndex: index,
      groupLabel: label,
      label,
      text,
      visualStep,
    })),
);

const sectionRanges = stepGroups.map(
  ({ label, steps: groupSteps }, groupIndex) => {
    const startStep =
      steps.findIndex((step) => step.groupIndex === groupIndex) + 1;
    const endStep = startStep + groupSteps.length - 1;

    return {
      groupIndex,
      label,
      startStep,
      endStep,
      steps: groupSteps,
    };
  },
);

const sectionRows = [
  [sectionRanges[0]],
  [sectionRanges[1]],
  [sectionRanges[2], sectionRanges[4]],
  [sectionRanges[3], sectionRanges[5]],
];

const getVisibleState = (
  displayStep: number,
  options?: { disablePreparationAnimations?: boolean },
): VisibleState => {
  const disablePreparationAnimations =
    options?.disablePreparationAnimations ?? false;

  return {
    proposition: displayStep === 3,
    baseLine: displayStep >= 1,
    pointC: displayStep >= 2,
    rotatingAE: displayStep === 4,
    completingSquare: displayStep === 5,
    outerSquare: displayStep >= 6,
    diagonal: displayStep >= 6,
    movingCF: displayStep === 7 && !disablePreparationAnimations,
    vertical:
      displayStep >= 8 || (disablePreparationAnimations && displayStep >= 7),
    movingHI: displayStep === 8 && !disablePreparationAnimations,
    horizontal:
      displayStep >= 9 || (disablePreparationAnimations && displayStep >= 8),
    parallelogramFills: displayStep === 19,
    rightAnglesV1: displayStep === 9 || displayStep === 20,
    rightAnglesV2: displayStep === 10,
    equalSides: displayStep === 11 || displayStep === 21,
    equalSegmentPairs: displayStep === 25,
    equalSegmentQuartet: displayStep === 26,
    labeledTopBisectedAngle: displayStep >= 12 && displayStep <= 14,
    unlabeledTopBisectedAngle: displayStep === 22,
    labeledLowerBisectedAngle: displayStep >= 13 && displayStep <= 14,
    unlabeledSquareBisectedAnglesAtB: displayStep === 22,
    unlabeledEqualAnglesAtBAndG: displayStep === 23,
    equalAngleQuartet: displayStep === 24,
    redSquare:
      displayStep === 14 ||
      displayStep === 17 ||
      displayStep === 18 ||
      displayStep === 27 ||
      displayStep >= 29,
    blueSquare:
      displayStep === 15 ||
      displayStep === 17 ||
      displayStep === 18 ||
      displayStep === 27 ||
      displayStep >= 29,
    greenRects:
      displayStep === 16 ||
      displayStep === 17 ||
      displayStep === 18 ||
      displayStep >= 28 ,
  };
};

const Box = styled.div`
  background: white;
  padding: clamp(0.75rem, 1.8vw, 1.15rem);
  border: 1px solid ${PANE_COLOR};
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.08);
  border-radius: 0.5rem;
  width: min(78rem, 100%);
  max-width: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  color: #333;
  box-sizing: border-box;
  overflow-x: hidden;
`;

const StepHeader = styled.div`
  text-align: center;
  margin-bottom: 0.2rem;
  width: 100%;
  display: flex;
  flex-direction: column;
  justify-content: center;
`;

const StepInstruction = styled.div`
  display: flex;
  flex-direction: column;
  font-size: clamp(0.92rem, 1.5vw, 1rem);
  color: #555;
  gap: 0.5rem;
  justify-content: center;
`;

const StepProgress = styled.div`
  font-style: normal;
  font-size: 0.85rem;
  font-weight: 600;
  color: #8b7351;
`;

const Sections = styled.div`
  width: 100%;
  margin-bottom: 0.7rem;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
`;

const SectionRow = styled.div<{ columns?: 1 | 2 }>`
  width: 100%;
  display: grid;
  grid-template-columns: ${({ columns }) =>
    columns === 2 ? "repeat(2, minmax(0, 1fr))" : "minmax(0, 1fr)"};
  gap: 0.5rem;

  @media only screen and (max-width: 900px) {
    grid-template-columns: 1fr;
  }
`;

const SectionCard = styled.div`
  border: 1px solid #e6dfd3;
  border-radius: 0.45rem;
  background: #fffdfa;
  overflow: hidden;
`;

const SectionInner = styled.div`
  padding: 0.55rem;
  display: flex;
  flex-direction: column;
  gap: 0.55rem;
`;

const HypothesisGrid = styled.div`
  width: 100%;
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.75rem;

  @media only screen and (max-width: 900px) {
    grid-template-columns: 1fr;
  }
`;

const HypothesisPanel = styled.div`
  padding: 0.65rem;
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
  text-align: left;
`;

const DiagramFrame = styled.div`
  width: 100%;
  flex: 1 1 auto;
  min-height: 0;
  display: flex;
  align-items: center;
  justify-content: center;
`;

const Button = styled.button`
  padding: 0.45rem 1rem;
  cursor: pointer;
  border: 1px solid #888;
  background: #fff;
  font-variant: small-caps;
  min-width: 5.5rem;
  border-radius: 0.25rem;
  color: ${SEA_COLOR};

  &:hover:not(:disabled) {
    background: #f9f9f9;
    border-color: #b08d57;
  }

  &:disabled {
    color: #ccc;
    border-color: #eee;
    cursor: not-allowed;
  }
`;

const SectionToggle = styled.button`
  width: 100%;
  padding: 0.7rem 0.9rem;
  cursor: pointer;
  border: 0;
  border-bottom: 1px solid #e6dfd3;
  background: transparent;
  color: #5e513f;
  display: flex;
  align-items: center;
  justify-content: space-between;
  text-align: left;
  font-size: 0.98rem;
  font-weight: 600;

  &:hover {
    background: #f7f1e7;
  }
`;

const SectionControls = styled.div`
  display: flex;
  gap: 0.5rem;
  justify-content: center;
  flex-wrap: wrap;
`;

const ViewModeRow = styled.div`
  display: flex;
  justify-content: center;
`;

const ViewModeGroup = styled.div`
  display: flex;
  align-items: center;
  gap: 0.9rem;
  flex-wrap: wrap;
  justify-content: center;
  color: #5e513f;
  font-size: 0.92rem;
`;

const ViewModeLabel = styled.label`
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  cursor: pointer;
`;

const SequentialSteps = styled.div`
  display: flex;
  flex-direction: column;
  gap: 0.8rem;
`;

const SequentialStep = styled.div`
  display: flex;
  flex-direction: column;
  gap: 0.45rem;
  padding-top: 0.15rem;
`;

export function HerigoneMar26Proposition() {
  const [selectedSteps, setSelectedSteps] = useState<Record<number, number>>(
    {},
  );
  const [sectionViewModes, setSectionViewModes] = useState<
    Record<number, SectionViewMode>
  >({});
  const [collapsedSections, setCollapsedSections] = useState<
    Record<number, boolean>
  >({});

  const handleCardWheel = (event: WheelEvent<HTMLDivElement>) => {
    const scrollContainer = event.currentTarget.closest(
      "[data-herigone-scroll-container='true']",
    );

    if (!(scrollContainer instanceof HTMLElement)) {
      return;
    }

    scrollContainer.scrollTop += event.deltaY;
  };

  const openSection = (groupIndex: number) => {
    setCollapsedSections((sections) => ({
      ...sections,
      [groupIndex]: false,
    }));
  };

  const goToStep = (groupIndex: number, stepNumber: number) => {
    openSection(groupIndex);
    setSelectedSteps((stepsBySection) => ({
      ...stepsBySection,
      [groupIndex]: stepNumber,
    }));
  };

  const getSectionDisplayStep = (groupIndex: number, stepNumber: number) => {
    if (groupIndex === 0) {
      return 2;
    }

    if (groupIndex === 1) {
      return 3;
    }

    return stepNumber;
  };

  const renderDiagramForStep = (
    label: string,
    visualStep: number,
    groupIndex?: number,
  ) => {
    const disablePreparationAnimations =
      groupIndex === 4 && (visualStep === 7 || visualStep === 8);
    const visible = getVisibleState(visualStep, {
      disablePreparationAnimations,
    });

    return visible.proposition ? (
      <PropositionDiagram key={`${label}-${visualStep}`} />
    ) : (
      <ConstructionDiagram key={`${label}-${visualStep}`} visible={visible} />
    );
  };

  const renderSectionCard = ({
    label,
    startStep,
    endStep,
    steps: groupSteps,
    groupIndex,
  }: (typeof sectionRanges)[number]) => {
    const isCollapsed = collapsedSections[groupIndex] ?? false;
    const viewMode = sectionViewModes[groupIndex] ?? "step-by-step";
    const isStepByStep = viewMode === "step-by-step";
    const selectedStepNumber = selectedSteps[groupIndex] ?? 0;
    const sectionStepNumber =
      selectedStepNumber >= startStep && selectedStepNumber <= endStep
        ? selectedStepNumber
        : 0;
    const sectionStepData =
      sectionStepNumber >= startStep && sectionStepNumber <= endStep
        ? steps[sectionStepNumber - 1]
        : null;
    const sectionDisplayStep = getSectionDisplayStep(
      groupIndex,
      sectionStepData?.visualStep ?? 0,
    );
    const sectionInstruction = sectionStepData?.text;
    const canGoToPreviousStep =
      sectionStepNumber >= startStep && sectionStepNumber <= endStep
        ? sectionStepNumber > startStep
        : false;
    const canGoToNextStep =
      sectionStepNumber >= startStep && sectionStepNumber <= endStep
        ? sectionStepNumber < endStep
        : true;

    return (
      <SectionCard key={label}>
        <SectionToggle
          type="button"
          onClick={() =>
            setCollapsedSections((sections) => ({
              ...sections,
              [groupIndex]: !isCollapsed,
            }))
          }
        >
          <span>{label}</span>
          <span>{isCollapsed ? "+" : "-"}</span>
        </SectionToggle>
        {!isCollapsed && (
          <SectionInner>
            {groupIndex === 0 ? (
              <HypothesisGrid>
                {groupSteps.map(({ text, visualStep }, stepIndex) => {
                  return (
                    <HypothesisPanel key={`${label}-${stepIndex + 1}`}>
                      <StepInstruction>
                        {stepIndex + 1}. {text}
                      </StepInstruction>
                      <DiagramFrame>
                        <ConstructionDiagram
                          key={`${label}-${visualStep}`}
                          visible={getVisibleState(visualStep)}
                        />
                      </DiagramFrame>
                    </HypothesisPanel>
                  );
                })}
              </HypothesisGrid>
            ) : groupIndex === 1 ? (
              <>
                <StepHeader>
                  <StepInstruction>
                    {groupSteps[0]?.text?.split("\n").map((line, index) => (
                      <div key={index}>{line}</div>
                    ))}
                  </StepInstruction>
                </StepHeader>

                <DiagramFrame>
                  <PropositionDiagram key={`${label}-3`} />
                </DiagramFrame>
              </>
            ) : (
              <>
                <ViewModeRow>
                  <ViewModeGroup role="radiogroup" aria-label={`${label} view`}>
                    <ViewModeLabel>
                      <input
                        type="radio"
                        name={`section-view-${groupIndex}`}
                        checked={isStepByStep}
                        onChange={() =>
                          setSectionViewModes((modes) => ({
                            ...modes,
                            [groupIndex]: "step-by-step",
                          }))
                        }
                      />
                      <span>Step by step</span>
                    </ViewModeLabel>
                    <ViewModeLabel>
                      <input
                        type="radio"
                        name={`section-view-${groupIndex}`}
                        checked={!isStepByStep}
                        onChange={() =>
                          setSectionViewModes((modes) => ({
                            ...modes,
                            [groupIndex]: "sequential",
                          }))
                        }
                      />
                      <span>Sequential</span>
                    </ViewModeLabel>
                  </ViewModeGroup>
                </ViewModeRow>

                {isStepByStep ? (
                  <>
                    <StepHeader>
                      <StepInstruction>
                        <StepProgress>
                          {sectionStepNumber >= startStep &&
                          sectionStepNumber <= endStep
                            ? `${sectionStepNumber - startStep + 1}/${groupSteps.length}`
                            : `0/${groupSteps.length}`}
                        </StepProgress>
                        {sectionInstruction
                          ? sectionInstruction
                              .split("\n")
                              .map((line, index) => (
                                <div key={index}>{line}</div>
                              ))
                          : "\u00A0"}
                      </StepInstruction>
                    </StepHeader>

                    <DiagramFrame>
                      {renderDiagramForStep(
                        label,
                        sectionDisplayStep,
                        groupIndex,
                      )}
                    </DiagramFrame>

                    <SectionControls>
                      <Button
                        onClick={() =>
                          goToStep(
                            groupIndex,
                            Math.max(sectionStepNumber - 1, startStep),
                          )
                        }
                        disabled={!canGoToPreviousStep}
                      >
                        Previous
                      </Button>
                      <Button
                        onClick={() =>
                          goToStep(
                            groupIndex,
                            sectionStepNumber >= startStep &&
                              sectionStepNumber <= endStep
                              ? Math.min(sectionStepNumber + 1, endStep)
                              : startStep,
                          )
                        }
                        disabled={!canGoToNextStep}
                      >
                        Next
                      </Button>
                    </SectionControls>
                  </>
                ) : (
                  <SequentialSteps>
                    {groupSteps.map(({ text, visualStep }, stepIndex) => (
                      <SequentialStep key={`${label}-${visualStep}`}>
                        <StepHeader>
                          <StepInstruction>
                            <StepProgress>
                              {stepIndex + 1}/{groupSteps.length}
                            </StepProgress>
                            {text.split("\n").map((line, index) => (
                              <div key={index}>{line}</div>
                            ))}
                          </StepInstruction>
                        </StepHeader>
                        <DiagramFrame>
                          {renderDiagramForStep(
                            label,
                            getSectionDisplayStep(groupIndex, visualStep),
                            groupIndex,
                          )}
                        </DiagramFrame>
                      </SequentialStep>
                    ))}
                  </SequentialSteps>
                )}
              </>
            )}
          </SectionInner>
        )}
      </SectionCard>
    );
  };

  return (
    <Box onWheel={handleCardWheel}>
      <Sections>
        {sectionRows.map((row, rowIndex) => (
          <SectionRow
            key={`section-row-${rowIndex + 1}`}
            columns={row.length === 2 ? 2 : 1}
          >
            {row.map((section) => renderSectionCard(section))}
          </SectionRow>
        ))}
      </Sections>
    </Box>
  );
}
