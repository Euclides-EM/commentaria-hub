import styled from "@emotion/styled";

const Diagram = styled.svg`
  display: block;
  background: #fff;
  border: 1px solid #eee;
  width: min(72rem, calc(100vw - 3rem));
  height: min(24rem, 34vh, calc((100vw - 3rem) * 0.33));
  max-width: 100%;
  overflow: visible;
  margin-top: 0.3rem;
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
        stroke="black"
        strokeWidth="2"
      />
      <line x1="24" y1="348" x2="324" y2="348" stroke="black" strokeWidth="2" />
      <circle cx="224" cy="348" r="4" fill="#c00" />
      <Label x="18" y="34" textAnchor="end">
        A
      </Label>
      <Label x="330" y="34" textAnchor="start">
        B
      </Label>
      <Label x="18" y="372" textAnchor="end">
        A
      </Label>
      <Label x="224" y="372" textAnchor="middle" fill="#c00">
        C
      </Label>
      <Label x="330" y="372" textAnchor="start">
        B
      </Label>

      <rect
        x="450"
        y="90"
        width="200"
        height="200"
        fill="rgba(204,36,29,0.12)"
        stroke="#cc241d"
        strokeWidth="2"
      />
      <rect
        x="780"
        y="140"
        width="100"
        height="100"
        fill="rgba(38,139,210,0.12)"
        stroke="#268bd2"
        strokeWidth="2"
      />
      <rect
        x="990"
        y="90"
        width="100"
        height="200"
        fill="rgba(133,153,0,0.14)"
        stroke="#859900"
        strokeWidth="2"
      />
      <rect
        x="1170"
        y="90"
        width="100"
        height="200"
        fill="rgba(133,153,0,0.14)"
        stroke="#859900"
        strokeWidth="2"
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

      <Label x="444" y="84" textAnchor="end">
        C
      </Label>
      <Label x="656" y="84" textAnchor="start" fill="#c00">
        A
      </Label>
      <Label x="444" y="312" textAnchor="end">
        A
      </Label>
      <Label x="656" y="312" textAnchor="start" fill="#c00">
        C
      </Label>

      <Label x="774" y="134" textAnchor="end" fill="#c00">
        B
      </Label>
      <Label x="886" y="134" textAnchor="start">
        C
      </Label>
      <Label x="774" y="252" textAnchor="end" fill="#c00">
        C
      </Label>
      <Label x="886" y="252" textAnchor="start">
        B
      </Label>

      <Label x="984" y="84" textAnchor="end">
        A
      </Label>
      <Label x="984" y="312" textAnchor="end">
        C
      </Label>
      <Label x="1096" y="312" textAnchor="start">
        B
      </Label>

      <Label x="1164" y="84" textAnchor="end">
        A
      </Label>
      <Label x="1164" y="312" textAnchor="end">
        C
      </Label>
      <Label x="1276" y="312" textAnchor="start">
        B
      </Label>

      <text x="660" y="24" fontSize="17" textAnchor="middle" fill="#666">
        □ AB = □ AC + □ CB + 2 ▭ ACB
      </text>
    </Diagram>
  );
}
