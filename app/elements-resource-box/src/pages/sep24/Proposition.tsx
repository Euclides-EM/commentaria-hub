import { useEffect, useState } from "react";
import styled from "@emotion/styled";
import { MARKER_5, PANE_COLOR, SEA_COLOR } from "../../utils/colors";
import { withAppBasePath } from "../../utils/basePath";

type PropositionProps = {
  title: string;
  description: string[];
  stepImagePrefix: string;
  stepsToDescriptionIndex: Record<number, number>;
};

const Title = styled.div`
  font-size: 1.5rem;
  font-weight: bold;
  color: ${PANE_COLOR};

  @media only screen and (max-width: 500px) {
    font-size: 1.2rem;
  }
`;

const StepTitle = styled.div<{ selected: boolean }>`
  text-align: start;
  color: ${(props) => (props.selected ? MARKER_5 : "white")};
  line-height: 1.35rem;
`;

const Descriptions = styled.div`
  font-style: italic;
`;

const ButtonsRow = styled.div`
  display: flex;
  gap: 0.9rem;
`;

const Button = styled.button<{ hide: boolean }>`
  opacity: ${(props) => (props.hide ? 0 : 1)};
  pointer-events: ${(props) => (props.hide ? "none" : "auto")};
  width: 6rem;
  border-radius: 0.5rem;
  background-color: ${PANE_COLOR};
  color: ${SEA_COLOR};
  cursor: pointer;
`;

const Image = styled.img`
  max-height: min(15rem, 32vh);
  height: auto;
  max-width: min(94vw, 38rem);
`;

export function Proposition({
  title,
  stepImagePrefix,
  description,
  stepsToDescriptionIndex,
}: PropositionProps) {
  const steps = Object.keys(stepsToDescriptionIndex);
  const [step, setStep] = useState(0);
  const [playing, setPlaying] = useState(true);

  const reset = () => setStep(0);

  useEffect(() => {
    if (!playing) {
      return;
    }
    const interval = setInterval(() => {
      setStep((currentStep) => {
        if (currentStep === steps.length - 1) {
          setPlaying(false);
          clearInterval(interval);
          return currentStep;
        }
        return currentStep + 1;
      });
    }, 1000);
    return () => clearInterval(interval);
  }, [playing, steps.length]);

  return (
    <>
      <Title>{title}</Title>
      <Descriptions>
        {description.map((text, index) => (
          <StepTitle
            key={text}
            selected={index === stepsToDescriptionIndex[Number(steps[step])]}
          >
            {text}
          </StepTitle>
        ))}
      </Descriptions>
      <div>
        <Image
          src={withAppBasePath(`${stepImagePrefix}${steps[step]}.png`)}
          alt=""
        />
      </div>
      <ButtonsRow>
        <Button hide={false} onClick={reset}>
          Reset
        </Button>
        <Button
          hide={step === steps.length - 1}
          onClick={() => setPlaying((value) => !value)}
        >
          {playing ? "Pause" : "Play"}
        </Button>
      </ButtonsRow>
    </>
  );
}
