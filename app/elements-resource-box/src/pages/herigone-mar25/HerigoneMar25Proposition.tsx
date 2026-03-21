import { useMemo, useState } from "react";
import styled from "@emotion/styled";
import { PANE_COLOR, SEA_COLOR } from "../../utils/colors";
import { ConstructionDiagram } from "./diagrams/ConstructionDiagram";
import { PropositionDiagram } from "./diagrams/PropositionDiagram";

type Step = {
  label: string;
  text: string;
};

const steps: Step[] = [
  { label: "I. The Hypothesis", text: "Draw the straight line AB." },
  { label: "I. The Hypothesis", text: "Divide line AB at point C." },
  {
    label: "I. The Hypothesis",
    text: "Required to demonstrate: the area of a square with side AB equals the area of square AC plus the area of square CB plus 2 times the area of the rectangle made up of AC and CB.",
  },
  {
    label: "II. The Preparation",
    text: "Describe the square ADEB on the whole line AB.",
  },
  {
    label: "II. The Preparation",
    text: "Extend from E to add point D and complete the square.",
  },
  { label: "II. The Preparation", text: "Draw the diagonal EB." },
  {
    label: "II. The Preparation",
    text: "Draw CF parallel to AD; it intersects the diagonal at G.",
  },
  {
    label: "II. The Preparation",
    text: "Through G, draw HI parallel to AB.",
  },
  {
    label: "III. The Demonstration",
    text: "Angles AED, ABD, EHG, EFG, and HGF are right angles.",
  },
  {
    label: "III. The Demonstration",
    text: "Side AE is equal to side AB.",
  },
  {
    label: "III. The Demonstration",
    text: "The diagonal EB bisects the right angle AED, making angles AEB and DEB 45°.",
  },
  {
    label: "III. The Demonstration",
    text: "The diagonal EB bisects the right angle HGF, making angles HGE and FGE 45°.",
  },
  {
    label: "III. The Demonstration",
    text: "Since angle EHG is 90° and the other two are 45°, side HG must equal HE. EHGF is a square.",
  },
  {
    label: "III. The Demonstration",
    text: "Similarly, CGIB is also a square on the other segment.",
  },
  {
    label: "III. The Demonstration",
    text: "The remaining spaces are rectangles. Thus, the square on the whole equals the squares on the parts plus twice their rectangle. QED.",
  },
];

const quickLinks = [
  {
    label: "Hypothesis",
    step: steps.findIndex(({ label }) => label.includes("Hypothesis")) + 1,
  },
  {
    label: "Preparation",
    step: steps.findIndex(({ label }) => label.includes("Preparation")) + 1,
  },
  {
    label: "Demonstration",
    step: steps.findIndex(({ label }) => label.includes("Demonstration")) + 1,
  },
];

const Box = styled.div`
  background: white;
  padding: clamp(0.75rem, 1.8vw, 1.15rem);
  border: 1px solid ${PANE_COLOR};
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.08);
  border-radius: 0.5rem;
  width: min(78rem, calc(100vw - 1.5rem));
  max-height: min(86vh, 86dvh);
  display: flex;
  flex-direction: column;
  align-items: center;
  color: #333;
  box-sizing: border-box;
  overflow-x: hidden;
  overflow-y: auto;
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
`;

const QuickLinks = styled.div`
  width: 100%;
  margin-bottom: 0.7rem;
  display: flex;
  gap: 0.2rem;
  justify-content: center;
  border-bottom: 1px solid #e6dfd3;
`;

const Controls = styled.div`
  margin-top: 0.8rem;
  display: flex;
  gap: 0.75rem;
  flex-wrap: wrap;
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
`;

const IntroText = "Click 'Next' to begin the flow.";

export function HerigoneMar25Proposition() {
  const [currentStep, setCurrentStep] = useState(0);
  const stepData = currentStep === 0 ? null : steps[currentStep - 1];
  const displayStep = currentStep;
  const activeSection = stepData?.label ?? null;
  const stageTitle = useMemo(() => {
    if (!stepData) {
      return "Introduction";
    }

    const phaseSteps = steps.filter(({ label }) => label === stepData.label);
    const phaseIndex =
      phaseSteps.findIndex(({ text }) => text === stepData.text) + 1;

    return `${stepData.label} (${phaseIndex}/${phaseSteps.length})`;
  }, [stepData]);

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
      rightAngles: displayStep === 9,
      equalSides: displayStep === 10,
      topBisectedAngle: displayStep >= 11 && displayStep <= 13,
      lowerBisectedAngle: displayStep >= 12 && displayStep <= 13,
      redSquare: displayStep >= 13,
      blueSquare: displayStep >= 14,
      greenRects: displayStep >= 15,
    }),
    [displayStep],
  );

  return (
    <Box>
      <QuickLinks>
        {quickLinks.map(({ label, step }) => {
          const targetLabel = steps[step - 1]?.label ?? "";
          const active = activeSection === targetLabel;

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

      {displayStep === 3 ? (
        <PropositionDiagram key={displayStep} />
      ) : (
        <ConstructionDiagram key={displayStep} visible={visible} />
      )}

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
