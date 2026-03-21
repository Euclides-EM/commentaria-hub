import { useMemo, useState } from "react";
import styled from "@emotion/styled";
import { PANE_COLOR, SEA_COLOR } from "../../utils/colors";
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
  tabLabel: string;
  label: string;
  steps: Array<{
    text: string;
    visualStep: number;
  }>;
};

const stepGroups: StepGroup[] = [
  {
    tabLabel: "Hypothesis",
    label: "I. The Hypothesis",
    steps: [
      { text: "Draw the straight line AB.", visualStep: 1 },
      { text: "Divide line AB at point C.", visualStep: 2 },
      {
        text: "Required to demonstrate: the area of a square with side AB equals the area of square AC plus the area of square CB plus 2 times the area of the rectangle made up of AC and CB.",
        visualStep: 3,
      },
    ],
  },
  {
    tabLabel: "Preparation",
    label: "II. The Preparation",
    steps: [
      { text: "Describe the square ADEB on the whole line AB.", visualStep: 4 },
      {
        text: "Extend from E to add point D and complete the square.",
        visualStep: 5,
      },
      { text: "Draw the diagonal EB.", visualStep: 6 },
      {
        text: "Draw CF parallel to AD; it intersects the diagonal at G.",
        visualStep: 7,
      },
      { text: "Through G, draw HI parallel to AB.", visualStep: 8 },
    ],
  },
  {
    tabLabel: "Demonstration (1634)",
    label: "III. The Demonstration (1634)",
    steps: [
      {
        text: "Angles AED, ABD, EHG, EFG, and HGF are right angles.",
        visualStep: 9,
      },
      { text: "Side AE is equal to side AB.", visualStep: 10 },
      {
        text: "The diagonal EB bisects the right angle AED, making angles AEB and DEB 45°.",
        visualStep: 11,
      },
      {
        text: "The diagonal EB bisects the right angle HGF, making angles HGE and FGE 45°.",
        visualStep: 12,
      },
      {
        text: "Since angle EHG is 90° and the other two are 45°, side HG must equal HE. EHGF is a square.",
        visualStep: 13,
      },
      {
        text: "Similarly, CGIB is also a square on the other segment.",
        visualStep: 14,
      },
      {
        text: "The remaining spaces are rectangles. Thus, the square on the whole equals the squares on the parts plus twice their rectangle. QED.",
        visualStep: 15,
      },
    ],
  },
  {
    tabLabel: "Demonstration (1639)",
    label: "III. The Demonstration (1639)",
    steps: [
      {
        text: "AG, HF, CI, and GD are parallelograms.",
        visualStep: 16,
      },
      {
        text: "Angles AED, ABD, EHG, EFG, and HGF are right angles.",
        visualStep: 17,
      },
      { text: "Side AE is equal to side AB.", visualStep: 18 },
      {
        text: "AEB, ABE, DEB, and DBE are all equal to each other, because the diagonal EB bisects the right angles of the large square.",
        visualStep: 19,
      },
      {
        text: "Similarly, CBG, CGB, HEG, and HGE are all equal to each other.",
        visualStep: 20,
      },
      {
        text: "BC equals CG, and GH equals HE.",
        visualStep: 21,
      },
      {
        text: "HE equals GF, HG equals EF, CB equals GI, and CG equals BI.",
        visualStep: 22,
      },
      {
        text: "GHFE is a square, and CBIG is a square.",
        visualStep: 23,
      },
      {
        text: "The rectangles AG and GD are equal to each other.",
        visualStep: 24,
      },
      {
        text: "Thus, as in 1634, the whole is obtained by adding everything together at the end.",
        visualStep: 25,
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
      label: `${label} (${index + 1}/${groupedSteps.length})`,
      text,
      visualStep,
    })),
);

const quickLinks = stepGroups.map(({ tabLabel, label }) => ({
  label: tabLabel,
  step: steps.findIndex(({ groupLabel }) => groupLabel === label) + 1,
  groupLabel: label,
}));

const Box = styled.div`
  background: white;
  padding: clamp(0.75rem, 1.8vw, 1.15rem);
  border: 1px solid ${PANE_COLOR};
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.08);
  border-radius: 0.5rem;
  width: min(78rem, 100%);
  max-width: 100%;
  max-height: min(86vh, 86dvh);
  display: flex;
  flex-direction: column;
  align-items: center;
  color: #333;
  box-sizing: border-box;
  overflow-x: hidden;
  overflow-y: auto;

  @media only screen and (max-height: 500px) and (orientation: landscape) {
    padding: 0.45rem 0.6rem;
    height: 100%;
    max-height: 100%;
    overflow-y: hidden;
  }

  @media only screen and (max-width: 500px) and (orientation: portrait) {
    padding: 0.45rem 0.6rem;
    width: 100%;
    height: 100%;
    max-height: 100%;
    overflow-y: hidden;
  }
`;

const StepHeader = styled.div`
  min-height: 6rem;
  text-align: center;
  margin-bottom: 0.4rem;
  width: 100%;
  border-bottom: 1px solid #eee;
  padding-bottom: 0.55rem;
  display: flex;
  flex-direction: column;
  justify-content: center;

  @media only screen and (max-height: 500px) and (orientation: landscape) {
    height: 4.2rem;
    min-height: 4.2rem;
    margin-bottom: 0.2rem;
    padding-bottom: 0.3rem;
  }

  @media only screen and (max-width: 500px) and (orientation: portrait) {
    height: 4.2rem;
    min-height: 4.2rem;
    margin-bottom: 0.2rem;
    padding-bottom: 0.3rem;
  }
`;

const StageTitle = styled.span`
  display: block;
  font-variant: small-caps;
  font-size: 0.9em;
  letter-spacing: 2px;
  color: #b08d57;
  margin-bottom: 0.5rem;
  font-weight: bold;
`;

const StepInstruction = styled.span`
  font-style: italic;
  font-size: clamp(0.92rem, 1.5vw, 1rem);
  line-height: 1.35;
  min-height: 2rem;
  color: #555;

  @media only screen and (max-height: 500px) and (orientation: landscape) {
    font-size: 0.88rem;
    line-height: 1.2;
    min-height: 1.5rem;
  }

  @media only screen and (max-width: 500px) and (orientation: portrait) {
    font-size: 0.88rem;
    line-height: 1.2;
    min-height: 1.5rem;
  }
`;

const QuickLinks = styled.div`
  width: 100%;
  margin-bottom: 0.7rem;
  display: flex;
  gap: 0.2rem;
  justify-content: center;
  border-bottom: 1px solid #e6dfd3;

  @media only screen and (max-height: 500px) and (orientation: landscape) {
    margin-bottom: 0.3rem;
  }

  @media only screen and (max-width: 500px) and (orientation: portrait) {
    margin-bottom: 0.3rem;
  }
`;

const Controls = styled.div`
  margin-top: 0.8rem;
  display: flex;
  gap: 0.75rem;
  flex-wrap: wrap;
  justify-content: center;
  flex: 0 0 auto;

  @media only screen and (max-height: 500px) and (orientation: landscape) {
    margin-top: 0.35rem;
    gap: 0.45rem;
  }

  @media only screen and (max-width: 500px) and (orientation: portrait) {
    margin-top: 0.35rem;
    gap: 0.45rem;
  }
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

  @media only screen and (max-height: 500px) and (orientation: landscape) {
    min-width: 4.7rem;
    padding: 0.35rem 0.7rem;
  }

  @media only screen and (max-width: 500px) and (orientation: portrait) {
    min-width: 4.7rem;
    padding: 0.35rem 0.7rem;
  }
`;

const QuickLinkTab = styled.button<{ active: boolean }>`
  padding: 0.5rem 0.9rem 0.45rem;
  cursor: pointer;
  border: 0;
  border-bottom: 2px solid
    ${({ active }) => (active ? "#b08d57" : "transparent")};
  background: ${({ active }) => (active ? "#fbf7f0" : "transparent")};
  font-variant: small-caps;
  letter-spacing: 0.04em;
  border-top-left-radius: 0.35rem;
  border-top-right-radius: 0.35rem;
  color: ${({ active }) => (active ? SEA_COLOR : "#7b6a52")};

  &:hover:not(:disabled) {
    background: #f7f1e7;
    color: ${SEA_COLOR};
  }

  &:disabled {
    cursor: default;
  }

  @media only screen and (max-height: 500px) and (orientation: landscape) {
    padding: 0.35rem 0.55rem 0.3rem;
    font-size: 0.86rem;
  }

  @media only screen and (max-width: 500px) and (orientation: portrait) {
    padding: 0.35rem 0.55rem 0.3rem;
    font-size: 0.86rem;
  }
`;

const IntroText = "Click 'Next' to begin the flow.";

export function HerigoneMar26Proposition() {
  const [currentStep, setCurrentStep] = useState(0);
  const stepData = currentStep === 0 ? null : steps[currentStep - 1];
  const displayStep = stepData?.visualStep ?? 0;
  const activeSection = stepData?.groupLabel ?? null;
  const stageTitle = stepData?.label ?? "Introduction";

  const visible = useMemo(
    () => ({
      proposition: displayStep === 3,
      baseLine: displayStep >= 1,
      pointC: displayStep >= 2,
      rotatingAE: displayStep === 4,
      completingSquare: displayStep === 5,
      outerSquare: displayStep >= 6,
      diagonal: displayStep >= 6,
      movingCF: displayStep === 7,
      vertical: displayStep >= 8,
      movingHI: displayStep === 8,
      horizontal: displayStep >= 9,
      parallelogramFills: displayStep === 16,
      rightAngles: displayStep === 9 || displayStep === 17,
      equalSides: displayStep === 10 || displayStep === 18,
      equalSegmentPairs: displayStep === 21,
      equalSegmentQuartet: displayStep === 22,
      labeledTopBisectedAngle: displayStep >= 11 && displayStep <= 13,
      unlabeledTopBisectedAngle: displayStep === 19,
      labeledLowerBisectedAngle: displayStep >= 12 && displayStep <= 13,
      unlabeledLowerBisectedAngle: displayStep === 20,
      unlabeledSquareBisectedAnglesAtB: displayStep === 19,
      unlabeledEqualAnglesAtBAndG: displayStep === 20,
      redSquare:
        (displayStep >= 13 && displayStep <= 15) ||
        displayStep === 23 ||
        displayStep === 25,
      blueSquare:
        (displayStep >= 14 && displayStep <= 15) ||
        displayStep === 23 ||
        displayStep === 25,
      greenRects: displayStep === 15 || displayStep >= 24,
      qed: displayStep === 15 || displayStep === 25,
    }),
    [displayStep],
  );

  return (
    <Box>
      <QuickLinks>
        {quickLinks.map(({ label, step, groupLabel }) => {
          const active = activeSection === groupLabel;

          return (
            <QuickLinkTab
              key={label}
              type="button"
              active={active}
              onClick={() => setCurrentStep(step)}
              disabled={step < 1 || active}
            >
              {label}
            </QuickLinkTab>
          );
        })}
      </QuickLinks>

      <StepHeader>
        <StageTitle>{stageTitle}</StageTitle>
        <StepInstruction>{stepData?.text ?? IntroText}</StepInstruction>
      </StepHeader>

      <DiagramFrame>
        {visible.proposition ? (
          <PropositionDiagram key={displayStep} />
        ) : (
          <ConstructionDiagram key={displayStep} visible={visible} />
        )}
      </DiagramFrame>

      <Controls>
        <Button
          onClick={() => setCurrentStep((step) => Math.max(step - 1, 0))}
          disabled={currentStep === 0}
        >
          Previous
        </Button>
        <Button
          onClick={() =>
            setCurrentStep((step) => Math.min(step + 1, steps.length))
          }
          disabled={currentStep === steps.length}
        >
          Next
        </Button>
        <Button onClick={() => setCurrentStep(0)}>Reset</Button>
      </Controls>
    </Box>
  );
}
