import styled from "@emotion/styled";

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
        stroke="black"
        strokeWidth="2"
      />
      <line
        x1="24"
        y1="340"
        x2="24"
        y2="140"
        stroke="#cc241d"
        strokeWidth="3"
      />
      <line x1="24" y1="140" x2="24" y2="40" stroke="#268bd2" strokeWidth="3" />
      <line
        x1="24"
        y1="340"
        x2="224"
        y2="340"
        stroke="#cc241d"
        strokeWidth="3"
      />
      <line
        x1="224"
        y1="340"
        x2="324"
        y2="340"
        stroke="#268bd2"
        strokeWidth="3"
      />
      <circle cx="224" cy="340" r="4" fill="#c00" />
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
      <line x1="450" y1="90" x2="650" y2="90" stroke="black" strokeWidth="2" />
      <line x1="650" y1="90" x2="650" y2="290" stroke="black" strokeWidth="2" />
      <line
        x1="450"
        y1="290"
        x2="650"
        y2="290"
        stroke="#cc241d"
        strokeWidth="3"
      />
      <line
        x1="450"
        y1="290"
        x2="450"
        y2="90"
        stroke="#cc241d"
        strokeWidth="3"
      />
      <rect x="780" y="140" width="100" height="100" fill="none" />
      <line
        x1="780"
        y1="140"
        x2="880"
        y2="140"
        stroke="black"
        strokeWidth="2"
      />
      <line
        x1="880"
        y1="140"
        x2="880"
        y2="240"
        stroke="black"
        strokeWidth="2"
      />
      <line
        x1="780"
        y1="240"
        x2="880"
        y2="240"
        stroke="#268bd2"
        strokeWidth="3"
      />
      <line
        x1="780"
        y1="240"
        x2="780"
        y2="140"
        stroke="#268bd2"
        strokeWidth="3"
      />
      <rect x="990" y="90" width="100" height="200" fill="none" />
      <line x1="990" y1="90" x2="1090" y2="90" stroke="black" strokeWidth="2" />
      <line
        x1="1090"
        y1="90"
        x2="1090"
        y2="290"
        stroke="black"
        strokeWidth="2"
      />
      <line
        x1="990"
        y1="90"
        x2="990"
        y2="290"
        stroke="#cc241d"
        strokeWidth="2"
      />
      <line
        x1="990"
        y1="290"
        x2="1090"
        y2="290"
        stroke="#268bd2"
        strokeWidth="2"
      />
      <rect x="1170" y="90" width="100" height="200" fill="none" />
      <line
        x1="1170"
        y1="90"
        x2="1270"
        y2="90"
        stroke="black"
        strokeWidth="2"
      />
      <line
        x1="1270"
        y1="90"
        x2="1270"
        y2="290"
        stroke="black"
        strokeWidth="2"
      />
      <line
        x1="1170"
        y1="90"
        x2="1170"
        y2="290"
        stroke="#cc241d"
        strokeWidth="2"
      />
      <line
        x1="1170"
        y1="290"
        x2="1270"
        y2="290"
        stroke="#268bd2"
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
    </Diagram>
  );
}
