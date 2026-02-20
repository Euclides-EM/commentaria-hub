import styled from "@emotion/styled";
import { Row } from "../../common.ts";
import { useCallback, useEffect, useRef, useState } from "react";
import { RANGE_FILL, SEA_COLOR } from "../../../utils/colors.ts";
import { MAX_YEAR, MIN_YEAR } from "../../../types";

export type RangeSliderProps = {
  className?: string;
  name?: string;
  value: [number, number];
  min: number;
  max: number;
  onChange: (value: [number, number]) => void;
};

const SliderContainer = styled.div`
  display: flex;
  position: relative;
  width: 20rem;
  height: 1rem;
  align-items: center;
`;

const SliderTrack = styled.div`
  position: absolute;
  width: 100%;
  height: 0.25rem;
  background-color: #ddd;
  border-radius: 0.125rem;
`;

const SliderRange = styled.div<{
  left: number;
  width: number;
  min: number;
  max: number;
}>`
  position: absolute;
  left: ${({ left, min, max }) => ((left - min) / (max - min)) * 100}%;
  width: ${({ width, min, max }) => (width / (max - min)) * 100}%;
  height: 0.25rem;
  background-color: ${RANGE_FILL};
  border-radius: 0.125rem;
`;

const SliderInput = styled.input`
  position: absolute;
  width: 100%;
  margin: 0;
  background-color: transparent;
  -webkit-appearance: none;
  appearance: none;
  pointer-events: none;

  &::-webkit-slider-thumb {
    -webkit-appearance: none;
    height: 1rem;
    width: 1rem;
    border-radius: 50%;
    background: #ddd;
    border: 0.125rem solid ${RANGE_FILL};
    cursor: pointer;
    pointer-events: auto;
  }

  &::-webkit-slider-thumb:active {
    transform: scale(1.5);
  }
`;

const MinInput = styled(SliderInput)``;

const MaxInput = styled(SliderInput)``;

const ValueInput = styled.input`
  margin: 0 0.5rem;
  width: 4rem;
  height: 2rem;
  text-align: center;
  border: 1px solid ${SEA_COLOR};
  border-radius: 4px;
  padding: 2px 4px;
  background-color: white;
  color: ${RANGE_FILL};

  &::-webkit-inner-spin-button,
  &::-webkit-outer-spin-button {
    -webkit-appearance: none;
    margin: 0;
  }

  &[type="number"] {
    -moz-appearance: textfield;
  }
`;

export const RangeSlider = ({
  className,
  name,
  value,
  min,
  max,
  onChange,
}: RangeSliderProps) => {
  const resolvedMin = Number.isFinite(min) ? min : MIN_YEAR;
  const resolvedMax = Number.isFinite(max) ? max : MAX_YEAR;
  const initialMin = Math.max(
    resolvedMin,
    Number.isFinite(value?.[0]) ? value[0] : resolvedMin,
  );
  const initialMax = Math.min(
    resolvedMax,
    Number.isFinite(value?.[1]) ? value[1] : resolvedMax,
  );

  const [minInputValue, setMinInputValue] = useState(initialMin.toString());
  const [maxInputValue, setMaxInputValue] = useState(initialMax.toString());
  const [localValue, setLocalValue] = useState<[number, number]>([
    initialMin,
    initialMax,
  ]);
  const localValueRef = useRef<[number, number]>([initialMin, initialMax]);
  const prevValueRef = useRef<[number, number]>([initialMin, initialMax]);
  const pendingChangeRef = useRef<[number, number] | null>(null);
  const rafRef = useRef<number | null>(null);

  const emitChange = useCallback(
    (nextValue: [number, number], immediate = false) => {
      if (immediate) {
        if (rafRef.current !== null) {
          cancelAnimationFrame(rafRef.current);
          rafRef.current = null;
        }
        pendingChangeRef.current = null;
        onChange(nextValue);
        return;
      }
      pendingChangeRef.current = nextValue;
      if (rafRef.current === null) {
        rafRef.current = requestAnimationFrame(() => {
          rafRef.current = null;
          if (pendingChangeRef.current) {
            onChange(pendingChangeRef.current);
            pendingChangeRef.current = null;
          }
        });
      }
    },
    [onChange],
  );

  const commitChange = useCallback(() => {
    const nextValue = localValueRef.current;
    if (
      nextValue[0] !== prevValueRef.current[0] ||
      nextValue[1] !== prevValueRef.current[1]
    ) {
      emitChange(nextValue);
    }
  }, [emitChange]);

  useEffect(() => {
    const nextMin = Math.max(
      resolvedMin,
      Number.isFinite(value?.[0]) ? value[0] : resolvedMin,
    );
    const nextMax = Math.min(
      resolvedMax,
      Number.isFinite(value?.[1]) ? value[1] : resolvedMax,
    );
    if (
      nextMin !== prevValueRef.current[0] ||
      nextMax !== prevValueRef.current[1]
    ) {
      setLocalValue([nextMin, nextMax]);
      setMinInputValue(nextMin.toString());
      setMaxInputValue(nextMax.toString());
      localValueRef.current = [nextMin, nextMax];
      prevValueRef.current = [nextMin, nextMax];
    }
  }, [value, resolvedMin, resolvedMax]);

  useEffect(() => {
    return () => {
      if (rafRef.current !== null) {
        cancelAnimationFrame(rafRef.current);
      }
    };
  }, []);

  const commitMinInputValue = useCallback(
    (inputValue: string) => {
      const newMin = parseInt(inputValue);
      if (!isNaN(newMin)) {
        const clampedMin = Math.max(
          resolvedMin,
          Math.min(newMin, localValue[1]),
        );
        setMinInputValue(clampedMin.toString());
        if (clampedMin !== localValue[0]) {
          const nextValue: [number, number] = [clampedMin, localValue[1]];
          setLocalValue(nextValue);
          localValueRef.current = nextValue;
          emitChange(nextValue, true);
        }
      } else {
        setMinInputValue(localValue[0].toString());
      }
    },
    [emitChange, localValue, resolvedMin],
  );

  const commitMaxInputValue = useCallback(
    (inputValue: string) => {
      const newMax = parseInt(inputValue);
      if (!isNaN(newMax)) {
        const clampedMax = Math.min(
          resolvedMax,
          Math.max(newMax, localValue[0]),
        );
        setMaxInputValue(clampedMax.toString());
        if (clampedMax !== localValue[1]) {
          const nextValue: [number, number] = [localValue[0], clampedMax];
          setLocalValue(nextValue);
          localValueRef.current = nextValue;
          emitChange(nextValue, true);
        }
      } else {
        setMaxInputValue(localValue[1].toString());
      }
    },
    [emitChange, localValue, resolvedMax],
  );

  return (
    <Row justifyStart noWrap noWrapAlsoOnMobile className={className}>
      {name && <div>{name}:</div>}
      <ValueInput
        type="number"
        min={resolvedMin}
        max={localValue[1]}
        value={minInputValue}
        onChange={(e) => setMinInputValue(e.target.value)}
        onBlur={() => commitMinInputValue(minInputValue)}
        onKeyDown={(e) => {
          if (e.key === "Enter") {
            commitMinInputValue(minInputValue);
          }
        }}
      />
      <SliderContainer>
        <SliderTrack id="range-slider-track" />
        <SliderRange
          left={Math.max(localValue[0], resolvedMin)}
          width={
            Math.min(localValue[1], resolvedMax) -
            Math.max(localValue[0], resolvedMin)
          }
          min={resolvedMin}
          max={resolvedMax}
        />
        <MinInput
          type="range"
          min={resolvedMin}
          max={resolvedMax}
          value={localValue[0]}
          onChange={(e) => {
            const newMin = parseInt(e.target.value);
            const clampedMin = Math.min(newMin, localValue[1]);
            if (clampedMin !== localValue[0]) {
              const nextValue: [number, number] = [clampedMin, localValue[1]];
              setLocalValue(nextValue);
              localValueRef.current = nextValue;
            }
          }}
          onMouseUp={commitChange}
          onTouchEnd={commitChange}
          onKeyUp={commitChange}
        />
        <MaxInput
          type="range"
          min={resolvedMin}
          max={resolvedMax}
          value={localValue[1]}
          onChange={(e) => {
            const newMax = parseInt(e.target.value);
            const clampedMax = Math.max(newMax, localValue[0]);
            if (clampedMax !== localValue[1]) {
              const nextValue: [number, number] = [localValue[0], clampedMax];
              setLocalValue(nextValue);
              localValueRef.current = nextValue;
            }
          }}
          onMouseUp={commitChange}
          onTouchEnd={commitChange}
          onKeyUp={commitChange}
        />
      </SliderContainer>
      <ValueInput
        type="number"
        min={localValue[0]}
        max={resolvedMax}
        value={maxInputValue}
        onChange={(e) => setMaxInputValue(e.target.value)}
        onBlur={() => commitMaxInputValue(maxInputValue)}
        onKeyDown={(e) => {
          if (e.key === "Enter") {
            commitMaxInputValue(maxInputValue);
          }
        }}
      />
    </Row>
  );
};
