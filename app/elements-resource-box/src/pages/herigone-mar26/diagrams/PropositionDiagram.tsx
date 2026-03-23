import styled from "@emotion/styled";

const neutralStroke = "#6b7280";
const firstAccentStroke = "#b91c1c";
const secondAccentStroke = "#1d4ed8";
const neutralStrokeWidth = 1.5;
const accentStrokeWidth = 4.25;
const accentInset = accentStrokeWidth / 2;
const accentStrokeLinecap = "round";

const Diagram = styled.svg`
  display: block;
  background: #fff;
  border: 1px solid #eee;
  width: min(72rem, 100%);
  height: min(24rem, 34vh, 33vw);
  max-width: 100%;
  overflow: visible;
  margin-top: 0.3rem;

  @media only screen and (max-height: 500px) and (orientation: landscape) {
    width: 100%;
    max-width: 100%;
    height: auto;
    aspect-ratio: 1320 / 380;
    margin-top: 0.1rem;
  }
`;

const Label = styled.text`
  font-size: 20px;
  font-weight: bold;
  font-family: serif;
  fill: #222;
`;

export function PropositionDiagram() {
  return (
    <Diagram
      viewBox="0 0 1320 380"
      preserveAspectRatio="xMidYMid meet"
      role="img"
      aria-label="Visual equation for the square on AB"
    >
      <rect
        x="24"
        y="40"
        width="300"
        height="300"
        fill="none"
        stroke={neutralStroke}
        strokeWidth={neutralStrokeWidth}
      />
      <line
        x1="24"
        y1={340 - accentInset}
        x2="24"
        y2="140"
        stroke={firstAccentStroke}
        strokeWidth={accentStrokeWidth}
        strokeLinecap={accentStrokeLinecap}
      />
      <line
        x1="24"
        y1="140"
        x2="24"
        y2={40 + accentInset}
        stroke={secondAccentStroke}
        strokeWidth={accentStrokeWidth}
        strokeLinecap={accentStrokeLinecap}
      />
      <line
        x1="24"
        y1="340"
        x2="224"
        y2="340"
        stroke={firstAccentStroke}
        strokeWidth={accentStrokeWidth}
        strokeLinecap={accentStrokeLinecap}
      />
      <line
        x1="224"
        y1="340"
        x2={324 - accentInset}
        y2="340"
        stroke={secondAccentStroke}
        strokeWidth={accentStrokeWidth}
        strokeLinecap={accentStrokeLinecap}
      />
      <circle cx="224" cy="340" r="4" fill={firstAccentStroke} />
      <Label x="18" y="362" textAnchor="end">
        A
      </Label>
      <Label x="224" y="362" textAnchor="middle" fill="#c00">
        C
      </Label>
      <Label x="330" y="362" textAnchor="start">
        B
      </Label>

      <rect x="450" y="90" width="200" height="200" fill="none" />
      <line
        x1="450"
        y1="90"
        x2="650"
        y2="90"
        stroke={neutralStroke}
        strokeWidth={neutralStrokeWidth}
      />
      <line
        x1="650"
        y1="90"
        x2="650"
        y2="290"
        stroke={neutralStroke}
        strokeWidth={neutralStrokeWidth}
      />
      <path
        d={`M 450 ${90 + accentInset} L 450 290 L ${650 - accentInset} 290`}
        fill="none"
        stroke={firstAccentStroke}
        strokeWidth={accentStrokeWidth}
        strokeLinecap={accentStrokeLinecap}
        strokeLinejoin="round"
      />
      <rect x="780" y="140" width="100" height="100" fill="none" />
      <line
        x1="780"
        y1="140"
        x2="880"
        y2="140"
        stroke={neutralStroke}
        strokeWidth={neutralStrokeWidth}
      />
      <line
        x1="880"
        y1="140"
        x2="880"
        y2="240"
        stroke={neutralStroke}
        strokeWidth={neutralStrokeWidth}
      />
      <path
        d={`M 780 ${140 + accentInset} L 780 240 L ${880 - accentInset} 240`}
        fill="none"
        stroke={secondAccentStroke}
        strokeWidth={accentStrokeWidth}
        strokeLinecap={accentStrokeLinecap}
        strokeLinejoin="round"
      />
      <rect x="990" y="90" width="100" height="200" fill="none" />
      <line
        x1="990"
        y1="90"
        x2="1090"
        y2="90"
        stroke={neutralStroke}
        strokeWidth={neutralStrokeWidth}
      />
      <line
        x1="1090"
        y1="90"
        x2="1090"
        y2="290"
        stroke={neutralStroke}
        strokeWidth={neutralStrokeWidth}
      />
      <line
        x1="990"
        y1={90 + accentInset}
        x2="990"
        y2="290"
        stroke={firstAccentStroke}
        strokeWidth={accentStrokeWidth}
        strokeLinecap={accentStrokeLinecap}
      />
      <line
        x1="990"
        y1="290"
        x2={1090 - accentInset}
        y2="290"
        stroke={secondAccentStroke}
        strokeWidth={accentStrokeWidth}
        strokeLinecap={accentStrokeLinecap}
      />
      <rect x="1170" y="90" width="100" height="200" fill="none" />
      <line
        x1="1170"
        y1="90"
        x2="1270"
        y2="90"
        stroke={neutralStroke}
        strokeWidth={neutralStrokeWidth}
      />
      <line
        x1="1270"
        y1="90"
        x2="1270"
        y2="290"
        stroke={neutralStroke}
        strokeWidth={neutralStrokeWidth}
      />
      <line
        x1="1170"
        y1={90 + accentInset}
        x2="1170"
        y2="290"
        stroke={firstAccentStroke}
        strokeWidth={accentStrokeWidth}
        strokeLinecap={accentStrokeLinecap}
      />
      <line
        x1="1170"
        y1="290"
        x2={1270 - accentInset}
        y2="290"
        stroke={secondAccentStroke}
        strokeWidth={accentStrokeWidth}
        strokeLinecap={accentStrokeLinecap}
      />

      <text x="388" y="196" fontSize="34" textAnchor="middle" fill="#666">
        =
      </text>
      <text x="716" y="196" fontSize="30" textAnchor="middle" fill="#666">
        +
      </text>
      <text x="936" y="196" fontSize="30" textAnchor="middle" fill="#666">
        +
      </text>
      <text x="1134" y="196" fontSize="30" textAnchor="middle" fill="#666">
        +
      </text>
    </Diagram>
  );
}
