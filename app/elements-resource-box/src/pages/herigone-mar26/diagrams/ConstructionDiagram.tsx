import { useCallback, useEffect, useRef } from "react";
import styled from "@emotion/styled";
import { type VisibleState } from "../visibleState";

type ConstructionVisibleState = Omit<VisibleState, "proposition" | "qed">;

const Diagram = styled.svg<{ $isReplayable: boolean }>`
  display: block;
  background: #fff;
  border: 1px solid #eee;
  width: min(72rem, 100%);
  height: min(24rem, 34vh, 33vw);
  max-width: 100%;
  overflow: visible;
  margin-top: 0.3rem;
  cursor: ${({ $isReplayable }) => ($isReplayable ? "pointer" : "default")};

  @media only screen and (max-height: 500px) and (orientation: landscape) {
    margin-top: 0.1rem;
    height: min(14rem, calc(100dvh - 10rem));
  }

  @media only screen and (max-width: 500px) and (orientation: portrait) {
    margin-top: 0.1rem;
    height: min(14rem, calc(100dvh - 10rem));
  }
`;

const Label = styled.text`
  font-size: 20px;
  font-weight: bold;
  font-family: serif;
  fill: #222;
`;

function RightAngle({
  x,
  y,
  fillPath,
  strokePath,
}: {
  x: string;
  y: string;
  fillPath: string;
  strokePath: string;
}) {
  return (
    <g transform={`translate(${x} ${y})`}>
      <path d={fillPath} fill="rgba(52, 122, 189, 0.18)" stroke="none" />
      <path
        d={strokePath}
        stroke="#347abd"
        strokeWidth="1.5"
        strokeLinecap="square"
        strokeLinejoin="miter"
        fill="none"
      />
    </g>
  );
}

export function ConstructionDiagram({
  visible,
}: {
  visible: ConstructionVisibleState;
}) {
  const timeoutIdsRef = useRef<number[]>([]);
  const rotatingAETransformRef = useRef<SVGAnimateTransformElement | null>(
    null,
  );
  const rotatingAESetRef = useRef<SVGSetElement | null>(null);
  const completingSquareTransformRef =
    useRef<SVGAnimateTransformElement | null>(null);
  const completingSquareTopRef = useRef<SVGSetElement | null>(null);
  const completingSquareRightRef = useRef<SVGSetElement | null>(null);
  const completingSquareLabelRef = useRef<SVGSetElement | null>(null);
  const movingCFTransformRef = useRef<SVGAnimateTransformElement | null>(null);
  const movingCFLabelFRef = useRef<SVGSetElement | null>(null);
  const movingCFLabelGRef = useRef<SVGSetElement | null>(null);
  const movingHITransformRef = useRef<SVGAnimateTransformElement | null>(null);
  const movingHILabelHRef = useRef<SVGSetElement | null>(null);
  const movingHILabelIRef = useRef<SVGSetElement | null>(null);
  const isReplayable =
    visible.rotatingAE ||
    visible.completingSquare ||
    visible.movingCF ||
    visible.movingHI;

  const clearReplayTimeouts = useCallback(() => {
    timeoutIdsRef.current.forEach((timeoutId) =>
      window.clearTimeout(timeoutId),
    );
    timeoutIdsRef.current = [];
  }, []);

  const replay = useCallback(() => {
    clearReplayTimeouts();

    if (visible.rotatingAE) {
      rotatingAETransformRef.current?.beginElement?.();
      timeoutIdsRef.current.push(
        window.setTimeout(() => {
          rotatingAESetRef.current?.beginElement?.();
        }, 1400),
      );
    }

    if (visible.completingSquare) {
      completingSquareTransformRef.current?.beginElement?.();
      timeoutIdsRef.current.push(
        window.setTimeout(() => {
          completingSquareTopRef.current?.beginElement?.();
          completingSquareRightRef.current?.beginElement?.();
          completingSquareLabelRef.current?.beginElement?.();
        }, 1400),
      );
    }

    if (visible.movingCF) {
      movingCFTransformRef.current?.beginElement?.();
      timeoutIdsRef.current.push(
        window.setTimeout(() => {
          movingCFLabelFRef.current?.beginElement?.();
          movingCFLabelGRef.current?.beginElement?.();
        }, 2000),
      );
    }

    if (visible.movingHI) {
      movingHITransformRef.current?.beginElement?.();
      timeoutIdsRef.current.push(
        window.setTimeout(() => {
          movingHILabelHRef.current?.beginElement?.();
          movingHILabelIRef.current?.beginElement?.();
        }, 2000),
      );
    }
  }, [
    clearReplayTimeouts,
    visible.completingSquare,
    visible.movingCF,
    visible.movingHI,
    visible.rotatingAE,
  ]);

  useEffect(() => {
    replay();

    return clearReplayTimeouts;
  }, [clearReplayTimeouts, replay]);

  return (
    <Diagram
      $isReplayable={isReplayable}
      viewBox="0 0 400 400"
      preserveAspectRatio="xMidYMid meet"
      role="img"
      aria-label="Interactive geometric proof"
      onClick={isReplayable ? replay : undefined}
    >
      {visible.baseLine && (
        <g>
          <line
            x1="50"
            y1="350"
            x2="350"
            y2="350"
            stroke="black"
            strokeWidth="2"
          />
          <Label x="45" y="375" textAnchor="end">
            A
          </Label>
          <Label x="355" y="375" textAnchor="start">
            B
          </Label>
        </g>
      )}
      {visible.pointC && (
        <g>
          <circle cx="250" cy="350" r="4" fill="#c00" />
          <Label x="250" y="375" textAnchor="middle" fill="#c00">
            C
          </Label>
        </g>
      )}
      {visible.rotatingAE && (
        <g>
          <line
            x1="50"
            y1="350"
            x2="350"
            y2="350"
            stroke="black"
            strokeWidth="2"
          >
            <animateTransform
              ref={rotatingAETransformRef}
              attributeName="transform"
              type="rotate"
              from="0 50 350"
              to="-90 50 350"
              begin="indefinite"
              dur="1.4s"
              fill="freeze"
            />
          </line>
          <Label x="45" y="45" textAnchor="end" visibility="hidden">
            E
            <set
              ref={rotatingAESetRef}
              attributeName="visibility"
              to="visible"
              begin="indefinite"
              fill="freeze"
            />
          </Label>
        </g>
      )}
      {visible.completingSquare && (
        <g>
          <line
            x1="50"
            y1="350"
            x2="50"
            y2="50"
            stroke="black"
            strokeWidth="2"
          />
          <line
            x1="50"
            y1="350"
            x2="350"
            y2="350"
            stroke="black"
            strokeWidth="2"
          />
          <line
            x1="50"
            y1="50"
            x2="50"
            y2="350"
            stroke="#000"
            strokeOpacity="1"
            strokeWidth="2.5"
            shapeRendering="crispEdges"
          >
            <animateTransform
              ref={completingSquareTransformRef}
              attributeName="transform"
              type="rotate"
              from="0 50 50"
              to="-90 50 50"
              begin="indefinite"
              dur="1.4s"
              fill="freeze"
            />
          </line>
          <line
            x1="50"
            y1="50"
            x2="350"
            y2="50"
            stroke="black"
            strokeWidth="2"
            visibility="hidden"
          >
            <set
              ref={completingSquareTopRef}
              attributeName="visibility"
              to="visible"
              begin="indefinite"
              fill="freeze"
            />
          </line>
          <line
            x1="350"
            y1="350"
            x2="350"
            y2="50"
            stroke="black"
            strokeWidth="2"
            visibility="hidden"
          >
            <set
              ref={completingSquareRightRef}
              attributeName="visibility"
              to="visible"
              begin="indefinite"
              fill="freeze"
            />
          </line>
          <Label x="45" y="45" textAnchor="end">
            E
          </Label>
          <Label x="355" y="45" textAnchor="start" visibility="hidden">
            D
            <set
              ref={completingSquareLabelRef}
              attributeName="visibility"
              to="visible"
              begin="indefinite"
              fill="freeze"
            />
          </Label>
        </g>
      )}
      {visible.outerSquare && (
        <g>
          <rect
            x="50"
            y="50"
            width="300"
            height="300"
            fill="none"
            stroke="black"
            strokeWidth="2"
          />
          <Label x="355" y="45" textAnchor="start">
            D
          </Label>
          <Label x="45" y="45" textAnchor="end">
            E
          </Label>
        </g>
      )}
      {visible.diagonal && (
        <line
          x1="50"
          y1="50"
          x2="350"
          y2="350"
          stroke="#999"
          strokeWidth="1.5"
          strokeDasharray="6,4"
        />
      )}
      {visible.movingCF && (
        <g>
          <line
            x1="50"
            y1="350"
            x2="50"
            y2="50"
            stroke="black"
            strokeWidth="1.5"
          >
            <animateTransform
              ref={movingCFTransformRef}
              attributeName="transform"
              type="translate"
              from="0 0"
              to="200 0"
              begin="indefinite"
              dur="2s"
              fill="freeze"
            />
          </line>
          <Label x="250" y="40" textAnchor="middle" visibility="hidden">
            F
            <set
              ref={movingCFLabelFRef}
              attributeName="visibility"
              to="visible"
              begin="indefinite"
              fill="freeze"
            />
          </Label>
          <Label x="265" y="245" visibility="hidden">
            G
            <set
              ref={movingCFLabelGRef}
              attributeName="visibility"
              to="visible"
              begin="indefinite"
              fill="freeze"
            />
          </Label>
        </g>
      )}
      {visible.vertical && (
        <g>
          <line
            x1="250"
            y1="350"
            x2="250"
            y2="50"
            stroke="black"
            strokeWidth="1.5"
          />
          <Label x="250" y="40" textAnchor="middle">
            F
          </Label>
          <Label x="265" y="245">
            G
          </Label>
        </g>
      )}
      {visible.movingHI && (
        <g>
          <line
            x1="50"
            y1="350"
            x2="350"
            y2="350"
            stroke="black"
            strokeWidth="1.5"
          >
            <animateTransform
              ref={movingHITransformRef}
              attributeName="transform"
              type="translate"
              from="0 0"
              to="0 -100"
              begin="indefinite"
              dur="2s"
              fill="freeze"
            />
          </line>
          <Label x="40" y="255" textAnchor="end" visibility="hidden">
            H
            <set
              ref={movingHILabelHRef}
              attributeName="visibility"
              to="visible"
              begin="indefinite"
              fill="freeze"
            />
          </Label>
          <Label x="360" y="255" textAnchor="start" visibility="hidden">
            I
            <set
              ref={movingHILabelIRef}
              attributeName="visibility"
              to="visible"
              begin="indefinite"
              fill="freeze"
            />
          </Label>
        </g>
      )}
      {visible.horizontal && (
        <g>
          <line
            x1="50"
            y1="250"
            x2="350"
            y2="250"
            stroke="black"
            strokeWidth="1.5"
          />
          <Label x="40" y="255" textAnchor="end">
            H
          </Label>
          <Label x="360" y="255" textAnchor="start">
            I
          </Label>
        </g>
      )}
      {visible.parallelogramFills && (
        <g>
          <rect
            x="50"
            y="50"
            width="200"
            height="200"
            fill="rgba(196, 78, 82, 0.18)"
            stroke="#c44e52"
            strokeWidth="1.5"
          />
          <rect
            x="250"
            y="50"
            width="100"
            height="200"
            fill="rgba(129, 114, 179, 0.18)"
            stroke="#8172b3"
            strokeWidth="1.5"
          />
          <rect
            x="50"
            y="250"
            width="200"
            height="100"
            fill="rgba(85, 168, 104, 0.18)"
            stroke="#55a868"
            strokeWidth="1.5"
          />
          <rect
            x="250"
            y="250"
            width="100"
            height="100"
            fill="rgba(76, 114, 176, 0.18)"
            stroke="#4c72b0"
            strokeWidth="1.5"
          />
        </g>
      )}
      {visible.rightAnglesV1 && (
        <g>
          <RightAngle
            x="50"
            y="50"
            fillPath="M 0 16 L 0 0 L 16 0 L 16 16 Z"
            strokePath="M 16 0 L 16 16 M 16 16 L 0 16"
          />
          <RightAngle
            x="350"
            y="350"
            fillPath="M 0 -16 L 0 0 L -16 0 L -16 -16 Z"
            strokePath="M -16 0 L -16 -16 M -16 -16 L 0 -16"
          />
          <RightAngle
            x="50"
            y="350"
            fillPath="M 0 -16 L 0 0 L 16 0 L 16 -16 Z"
            strokePath="M 16 0 L 16 -16 M 16 -16 L 0 -16"
          />
          <RightAngle
            x="350"
            y="50"
            fillPath="M 0 16 L 0 0 L -16 0 L -16 16 Z"
            strokePath="M -16 0 L -16 16 M -16 16 L 0 16"
          />
        </g>
      )}
      {visible.rightAnglesV2 && (
        <g>
          <RightAngle
            x="50"
            y="250"
            fillPath="M 0 -16 L 0 0 L 16 0 L 16 -16 Z"
            strokePath="M 16 0 L 16 -16 M 16 -16 L 0 -16"
          />
          <RightAngle
            x="250"
            y="50"
            fillPath="M -16 0 L 0 0 L 0 16 L -16 16 Z"
            strokePath="M -16 0 L -16 16 M -16 16 L 0 16"
          />
          <RightAngle
            x="250"
            y="250"
            fillPath="M -16 0 L 0 0 L 0 -16 L -16 -16 Z"
            strokePath="M -16 0 L -16 -16 M -16 -16 L 0 -16"
          />
        </g>
      )}
      {visible.equalSides && (
        <g>
          <line
            x1="50"
            y1="350"
            x2="50"
            y2="50"
            stroke="#347abd"
            strokeWidth="4"
            strokeLinecap="round"
          />
          <line
            x1="50"
            y1="350"
            x2="350"
            y2="350"
            stroke="#347abd"
            strokeWidth="4"
            strokeLinecap="round"
          />
        </g>
      )}
      {visible.equalSegmentPairs && (
        <g>
          <line
            x1="250"
            y1="350"
            x2="350"
            y2="350"
            stroke="#347abd"
            strokeWidth="4"
            strokeLinecap="round"
          />
          <line
            x1="250"
            y1="350"
            x2="250"
            y2="250"
            stroke="#347abd"
            strokeWidth="4"
            strokeLinecap="round"
          />
          <line
            x1="50"
            y1="250"
            x2="250"
            y2="250"
            stroke="#d17a22"
            strokeWidth="4"
            strokeLinecap="round"
          />
          <line
            x1="50"
            y1="50"
            x2="50"
            y2="250"
            stroke="#d17a22"
            strokeWidth="4"
            strokeLinecap="round"
          />
        </g>
      )}
      {visible.equalSegmentQuartet && (
        <g>
          <line
            x1="50"
            y1="50"
            x2="50"
            y2="250"
            stroke="#d17a22"
            strokeWidth="4"
            strokeLinecap="round"
          />
          <line
            x1="250"
            y1="250"
            x2="250"
            y2="50"
            stroke="#d17a22"
            strokeWidth="4"
            strokeLinecap="round"
          />
          <line
            x1="50"
            y1="250"
            x2="250"
            y2="250"
            stroke="#2c7a7b"
            strokeWidth="4"
            strokeLinecap="round"
          />
          <line
            x1="50"
            y1="50"
            x2="250"
            y2="50"
            stroke="#2c7a7b"
            strokeWidth="4"
            strokeLinecap="round"
          />
          <line
            x1="250"
            y1="350"
            x2="350"
            y2="350"
            stroke="#347abd"
            strokeWidth="4"
            strokeLinecap="round"
          />
          <line
            x1="250"
            y1="250"
            x2="350"
            y2="250"
            stroke="#347abd"
            strokeWidth="4"
            strokeLinecap="round"
          />
          <line
            x1="250"
            y1="250"
            x2="250"
            y2="350"
            stroke="#b04a9f"
            strokeWidth="4"
            strokeLinecap="round"
          />
          <line
            x1="350"
            y1="250"
            x2="350"
            y2="350"
            stroke="#b04a9f"
            strokeWidth="4"
            strokeLinecap="round"
          />
        </g>
      )}
      {visible.unlabeledSquareBisectedAnglesAtB && (
        <g>
          <path
            d="M 350 350 L 332 350 A 18 18 0 0 1 337.27 337.27 Z"
            fill="rgba(255,0,0,0.40)"
          />
          <path
            d="M 350 350 L 337.27 337.27 A 18 18 0 0 1 350 332 Z"
            fill="rgba(255,0,0,0.16)"
          />
        </g>
      )}
      {visible.unlabeledEqualAnglesAtBAndG && (
        <g>
          <path
            d="M 350 350 L 332 350 A 18 18 0 0 1 337.27 337.27 Z"
            fill="rgba(0,255,0,0.40)"
          />
          <path
            d="M 250 250 L 228 250 A 22 22 0 0 1 234.44 234.44 Z"
            fill="rgba(0,255,0,0.40)"
          />
          <path
            d="M 250 250 L 250 272 A 22 22 0 0 0 265.56 265.56 Z"
            fill="rgba(255,0,0,0.40)"
          />
          <path
            d="M 250 250 L 272 250 A 22 22 0 0 1 265.56 265.56 Z"
            fill="rgba(0,0,255,0.40)"
          />
          <path
            d="M 50 50 L 68.38 68.38 A 26 26 0 0 1 50 76 Z"
            fill="rgba(255,0,0,0.40)"
          />
          <path
            d="M 50 50 L 76 50 A 26 26 0 0 1 68.38 68.38 Z"
            fill="rgba(0,0,255,0.40)"
          />
        </g>
      )}
      {visible.equalAngleQuartet && (
        <g>
          <path
            d="M 350 350 L 332 350 A 18 18 0 0 1 337.27 337.27 Z"
            fill="rgba(255,0,0,0.40)"
          />
          <path
            d="M 250 250 L 250 272 A 22 22 0 0 0 265.56 265.56 Z"
            fill="rgba(255,0,0,0.40)"
          />
          <path
            d="M 250 250 L 228 250 A 22 22 0 0 1 234.44 234.44 Z"
            fill="rgba(255,0,0,0.40)"
          />
          <path
            d="M 50 50 L 68.38 68.38 A 26 26 0 0 1 50 76 Z"
            fill="rgba(255,0,0,0.40)"
          />
        </g>
      )}
      {visible.labeledTopBisectedAngle && (
        <g>
          <path
            d="M 50 50 L 76 50 A 26 26 0 0 1 68.38 68.38 Z"
            fill="rgba(255,0,0,0.40)"
          />
          <path
            d="M 50 50 L 68.38 68.38 A 26 26 0 0 1 50 76 Z"
            fill="rgba(255,0,0,0.16)"
          />
          <text x="79" y="66" fontSize="11" fontStyle="italic" fill="#c00">
            45°
          </text>
          <text x="52" y="84" fontSize="11" fontStyle="italic" fill="#c00">
            45°
          </text>
        </g>
      )}
      {visible.unlabeledTopBisectedAngle && (
        <g>
          <path
            d="M 50 50 L 76 50 A 26 26 0 0 1 68.38 68.38 Z"
            fill="rgba(255,0,0,0.40)"
          />
          <path
            d="M 50 50 L 68.38 68.38 A 26 26 0 0 1 50 76 Z"
            fill="rgba(255,0,0,0.16)"
          />
        </g>
      )}
      {visible.labeledLowerBisectedAngle && (
        <g>
          <path
            d="M 250 250 L 228 250 A 22 22 0 0 1 234.44 234.44 Z"
            fill="rgba(255,0,0,0.40)"
          />
          <path
            d="M 250 250 L 234.44 234.44 A 22 22 0 0 1 250 228 Z"
            fill="rgba(255,0,0,0.16)"
          />
          <text x="208" y="240" fontSize="11" fontStyle="italic" fill="#c00">
            45°
          </text>
          <text x="228" y="221" fontSize="11" fontStyle="italic" fill="#c00">
            45°
          </text>
        </g>
      )}
      {visible.redSquare && (
        <rect
          x="50"
          y="50"
          width="200"
          height="200"
          fill="rgba(255,0,0,0.1)"
          stroke="red"
          strokeWidth="2"
        />
      )}
      {visible.blueSquare && (
        <rect
          x="250"
          y="250"
          width="100"
          height="100"
          fill="rgba(0,0,255,0.1)"
          stroke="blue"
          strokeWidth="2"
        />
      )}
      {visible.greenRects && (
        <g>
          <rect
            x="50"
            y="250"
            width="200"
            height="100"
            fill="rgba(0,255,0,0.1)"
            stroke="green"
            strokeWidth="1.5"
          />
          <rect
            x="250"
            y="50"
            width="100"
            height="200"
            fill="rgba(0,255,0,0.1)"
            stroke="green"
            strokeWidth="1.5"
          />
        </g>
      )}
    </Diagram>
  );
}
